package login

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/password"
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
	passwords        port.PasswordReader
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
	username, issues := user.NewUsername(input.Username)
	if !issues.IsEmpty() {
		return Output{}, ErrInvalidCredentials
	}
	plain, issues := password.NewPlain(input.Password)
	if !issues.IsEmpty() {
		return Output{}, ErrInvalidCredentials
	}

	usr, err := uc.users.FindByUsername(ctx, username)
	if err != nil {
		return Output{}, fmt.Errorf("find user by username failed: %w", err)
	}
	if usr == nil {
		return Output{}, ErrInvalidCredentials
	}

	pwd, err := uc.passwords.FindByUserID(ctx, usr.ID())
	if err != nil {
		return Output{}, fmt.Errorf("find password by user id failed: %w", err)
	}
	if pwd == nil {
		return Output{}, ErrInvalidCredentials
	}

	ok, err := uc.pwdChecker.Verify(plain, pwd.Hash())
	if err != nil {
		return Output{}, fmt.Errorf("password verification failed: %w", err)
	}
	if !ok {
		return Output{}, ErrInvalidCredentials
	}

	now := uc.clock.Now()

	sess, err := session.NewSession(session.SessionCreationParams{
		UserID:    usr.ID(),
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, fmt.Errorf("session creation failed: %w", err)
	}

	issuedAccess, err := uc.accessIssuer.Issue(port.AccessTokenPayload{
		UserID: usr.ID(),
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
		CreatedAt: now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		return Output{}, fmt.Errorf("refresh token creation failed: %w", err)
	}

	err = uc.uow.Do(ctx, func(deps port.UowDeps) error {
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
	if err != nil {
		return Output{}, fmt.Errorf("session and refresh token saving failed: %w", err)
	}

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
