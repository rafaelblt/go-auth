// Package login verifies a username and password, opens a session, and hands
// back both tokens. Every failure answers the same way, and a missing account
// costs a bcrypt comparison too, so neither the response nor its timing says
// whether an account exists.
//
// See docs/architecture/usecases/login.md.
package login

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
)

type Input struct {
	Username string
	Password string
}

type Output struct {
	UserID       string
	SessionID    string
	AccessToken  usecase.AccessTokenDTO
	RefreshToken usecase.RefreshTokenDTO
}

type Login struct {
	credentials  credential.Checker
	accessIssuer port.AccessTokenIssuer
	uow          port.UnitOfWork
	clock        port.Clock
	refreshTTL   time.Duration
}

func (uc Login) Execute(ctx context.Context, input Input) (Output, error) {
	username, plain, err := credential.Parse(input.Username, input.Password)
	if err != nil {
		return Output{}, err
	}

	usr, pwd, err := uc.credentials.Check(ctx, username, plain)
	if err != nil {
		return Output{}, err
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

	refreshToken, refreshSecret, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: sess.ID(),
		ParentID:  nil,
		CreatedAt: now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		return Output{}, fmt.Errorf("refresh token creation failed: %w", err)
	}

	err = uc.uow.Do(ctx, func(deps port.UowDeps) error {
		// The session is saved only while the password is still the one
		// checked above, and a change of it waits until it is saved, so the
		// change's revocation reaches it.
		//
		// See docs/development/decisions/0055-login-locks-the-password-it-checked.md.
		err := deps.PasswordWriter.LockHash(ctx, pwd)
		if err != nil {
			return fmt.Errorf("password writer lock hash failed: %w", err)
		}

		err = deps.SessionWriter.Add(ctx, sess)
		if err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Add(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("refresh token writer failed: %w", err)
		}

		return nil
	})
	if errors.Is(err, password.ErrHashChanged) {
		return Output{}, credential.ErrPasswordChanged
	}
	if err != nil {
		return Output{}, fmt.Errorf("session and refresh token saving failed: %w", err)
	}

	output := Output{
		UserID:       usr.ID().String(),
		SessionID:    sess.ID().String(),
		AccessToken:  usecase.MapAccessTokenIssuedToDTO(issuedAccess),
		RefreshToken: usecase.MapRefreshTokenToDTO(refreshToken, refreshSecret),
	}
	return output, nil
}
