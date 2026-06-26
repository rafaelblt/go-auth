package login

import (
	"context"
	"fmt"
	"time"

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
	users            port.UserReader
	credentials      port.CredentialReader
	pwdChecker       port.PasswordChecker
	accessIssuer     port.AccessTokenIssuer
	refreshGenerator port.RefreshTokenGenerator
	uow              port.UnitOfWork
	clock            port.Clock
	refreshTTL       time.Duration
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

	sess, err := session.NewSession(session.SessionCreationParams{
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

	generatedRefresh, err := uc.refreshGenerator.Generate()
	if err != nil {
		return Output{}, fmt.Errorf("refresh token generator failed: %w", err)
	}

	refreshToken, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: sess.ID(),
		Hash:      generatedRefresh.Hash,
		ParentID:  nil,
		IssuedAt:  now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		return Output{}, fmt.Errorf("refresh token creation failed: %w", err)
	}

	uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.SessionWriter.Add(ctx, sess)
		if err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Add(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("refresh token writer failed: %w", err)
		}

		return nil
	})

	output := Output{
		AccessToken: usecase.AccessTokenDTO{
			Value:     issuedAccess.Token.Value(),
			ExpiresAt: issuedAccess.ExpiresAt,
		},
		RefreshToken: usecase.RefreshTokenDTO{
			Value:     generatedRefresh.Raw,
			ExpiresAt: refreshToken.ExpiresAt(),
		},
	}
	return output, nil
}
