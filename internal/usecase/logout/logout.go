// Package logout ends the session of a refresh token. It revokes the session
// whatever the state of the token, and writes nothing about the token itself.
//
// See docs/architecture/usecases.md#logout.
package logout

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Logout struct {
	sessions port.SessionReader
	tokens   port.RefreshTokenReader
	uow      port.UnitOfWork
	clock    port.Clock
}

type Input struct {
	RefreshToken string
}

type Output struct {
	UserID    string
	SessionID string
	// ReuseDetected is set when the token was already spent and its session
	// still active: someone else may hold the token's successor.
	ReuseDetected bool
	// SessionAlreadyRevoked is set when something else revoked the session
	// first, and the logout changed nothing.
	SessionAlreadyRevoked bool
}

var ErrTokenInvalid = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"invalid token",
)

func (uc *Logout) Execute(ctx context.Context, in Input) (Output, error) {
	now := uc.clock.Now()

	token, err := uc.resolveToken(ctx, in.RefreshToken)
	if err != nil {
		return Output{}, err
	}

	sess, err := uc.getSessionOfToken(ctx, token)
	if err != nil {
		return Output{}, err
	}

	output := Output{
		UserID:    sess.UserID().String(),
		SessionID: sess.ID().String(),
	}
	if sess.IsRevoked() {
		output.SessionAlreadyRevoked = true
		return output, nil
	}

	revoked, err := uc.revoke(ctx, sess, now)
	if err != nil {
		return Output{}, err
	}
	if !revoked {
		output.SessionAlreadyRevoked = true
		return output, nil
	}

	output.ReuseDetected = token.IsUsed()
	return output, nil
}

func (uc *Logout) resolveToken(ctx context.Context, raw string) (*session.RefreshToken, error) {
	secret, err := session.ParseRefreshTokenSecret(raw)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	token, err := uc.tokens.FindByHash(ctx, secret.Hash())
	if err != nil {
		return nil, fmt.Errorf("find refresh token by hash failed: %w", err)
	}
	if token == nil {
		return nil, ErrTokenInvalid
	}
	return token, nil
}

func (uc *Logout) getSessionOfToken(ctx context.Context, token *session.RefreshToken) (*session.Session, error) {
	sess, err := uc.sessions.FindByID(ctx, token.SessionID())
	if err != nil {
		return nil, fmt.Errorf("find session by id failed: %w", err)
	}
	if sess == nil {
		// The session cleanup deleted the session, and the token with it,
		// after the token was read: the token is unknown from now on.
		//
		// See docs/architecture/usecases.md#logout.
		return nil, ErrTokenInvalid
	}
	return sess, nil
}

// revoke reports whether it revoked the session: not when another revocation
// got there first.
func (uc *Logout) revoke(ctx context.Context, sess *session.Session, now time.Time) (bool, error) {
	sess.Revoke(now)
	err := uc.uow.Do(ctx, func(deps port.UowDeps) error {
		if err := deps.SessionWriter.Revoke(ctx, sess); err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}
		return nil
	})
	if errors.Is(err, session.ErrSessionAlreadyRevoked) {
		// Another revocation committed after the session was read: the
		// session is left as it is, as when it is read revoked.
		//
		// See docs/architecture/usecases.md#logout.
		return false, nil
	}
	if errors.Is(err, session.ErrSessionNotFound) {
		// The session cleanup deleted the session, and the token with it,
		// after the session was read: the token is unknown from now on.
		//
		// See docs/architecture/usecases.md#logout.
		return false, ErrTokenInvalid
	}
	if err != nil {
		return false, fmt.Errorf("session revocation failed: %w", err)
	}
	return true, nil
}
