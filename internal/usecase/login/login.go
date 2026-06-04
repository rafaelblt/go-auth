package login

import (
	"context"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/user"
)

type Input struct {
	Username string
	Password string
}

type Output struct {
	AccessToken  usecase.AccessTokenDTO
	RefreshToken usecase.RefreshTokenDTO
}

type Login struct {
	users         port.UserReader
	credentials   port.CredentialReader
	pwdChecker    port.PasswordChecker
	accessIssuer  port.AccessTokenIssuer
	refreshIssuer port.RefreshTokenIssuer
	uow           port.UnitOfWork
	clock         port.Clock
}

var ErrInvalidCredentials = usecase.NewError(
	"INVALID_CREDENTIALS", usecase.ErrorKindUnauthorized,
)

func (uc Login) Execute(ctx context.Context, input Input) (Output, error) {
	username, err := user.NewUsername(input.Username)
	if err != nil {
		return Output{}, ErrInvalidCredentials
	}
	password, err := credential.NewPlainPassword(input.Password)
	if err != nil {
		return Output{}, ErrInvalidCredentials
	}

	user, err := uc.users.FindByUsername(ctx, username)
	if err != nil {
		return Output{}, fmt.Errorf("find user by username failed: %w", err)
	}
	if user == nil {
		return Output{}, ErrInvalidCredentials
	}

	credential, err := uc.credentials.FindByUserAndKind(ctx, user.ID(), credential.KindPassword)
	if err != nil {
		return Output{}, fmt.Errorf("find credential by user and kind failed: %w", err)
	}
	if credential == nil {
		return Output{}, ErrInvalidCredentials
	}

	ok, err := uc.pwdChecker.Verify(password, credential.Secret())
	if err != nil {
		return Output{}, fmt.Errorf("password verification failed: %w", err)
	}
	if !ok {
		return Output{}, ErrInvalidCredentials
	}

	now := uc.clock.Now()

	session, err := session.NewSession(session.SessionCreationParams{
		UserID:   user.ID(),
		IssuedAt: now,
	})
	if err != nil {
		return Output{}, fmt.Errorf("session creation failed: %w", err)
	}

	issuedAccess, err := uc.accessIssuer.Issue(port.AccessTokenPayload{
		UserID: user.ID(),
	})
	if err != nil {
		return Output{}, fmt.Errorf("access token issuer failed: %w", err)
	}

	issuedRefresh, err := uc.refreshIssuer.Issue(port.RefreshTokenPayload{
		UserID:    user.ID(),
		SessionID: session.ID(),
	})
	if err != nil {
		return Output{}, fmt.Errorf("access token issuer failed: %w", err)
	}

	uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.SessionWriter.Save(ctx, session)
		if err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Save(ctx, issuedRefresh.Token)
		if err != nil {
			return fmt.Errorf("refresh token writer failed: %w", err)
		}

		return nil
	})

	output := Output{
		AccessToken: usecase.MapAccessTokenIssuedToDTO(issuedAccess),
		RefreshToken: usecase.MapRefreshTokenIssuedToDTO(issuedRefresh),
	}
	return output, nil
}
