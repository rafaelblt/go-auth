// Package port declares the interfaces the inside of the application depends on,
// so that nothing in usecase or domain names a concrete adapter. The
// infrastructure implements them and is injected at startup.
//
// See docs/architecture/usecases.md#ports.
package port

import (
	"context"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

// Unit Of Work

type UnitOfWork interface {
	Do(ctx context.Context, fn func(deps UowDeps) error) error
}

type UowDeps struct {
	UserWriter         UserWriter
	PasswordWriter     PasswordWriter
	SessionWriter      SessionWriter
	RefreshTokenWriter RefreshTokenWriter
}

// User

type UserReader interface {
	FindByID(context.Context, user.ID) (*user.User, error)
	FindByUsername(context.Context, user.Username) (*user.User, error)
}

type UserExistsChecker interface {
	ExistsByUsername(context.Context, user.Username) (bool, error)
}

type UserWriter interface {
	// Add inserts a user. It applies only while the username is free, and
	// returns user.ErrUsernameAlreadyExists when another insert took the
	// username first.
	Add(context.Context, *user.User) error
}

// Password

type PasswordReader interface {
	FindByID(context.Context, password.ID) (*password.Password, error)
	FindByUserID(context.Context, user.ID) (*password.Password, error)
}

type PasswordWriter interface {
	Add(context.Context, *password.Password) error
	// UpdateHash persists a hash already changed in memory. It applies only
	// while the stored hash is still previous, and returns
	// password.ErrHashChanged when another change got there first.
	UpdateHash(ctx context.Context, pwd *password.Password, previous password.Hashed) error
	// LockHash keeps the stored hash from changing until the transaction ends.
	// It applies only while the stored hash is still pwd's, and returns
	// password.ErrHashChanged when a change got there first.
	LockHash(ctx context.Context, pwd *password.Password) error
}

// Session

type SessionReader interface {
	FindByID(context.Context, session.SessionID) (*session.Session, error)
}

type SessionWriter interface {
	Add(context.Context, *session.Session) error
	// Revoke persists the revocation of a session already revoked in memory.
	// It applies only while the stored session is active, and returns
	// session.ErrSessionAlreadyRevoked when another revocation got there
	// first, or session.ErrSessionNotFound when the session is no longer
	// stored, as when the session cleanup deleted it after it was read.
	Revoke(context.Context, *session.Session) error
	// RevokeAllByUserID revokes, at revokedAt, every session of the user that
	// is not revoked yet. A session already revoked keeps its time, and a user
	// with no session to revoke is not an error.
	RevokeAllByUserID(ctx context.Context, userID user.ID, revokedAt time.Time) error
}

type SessionCleaner interface {
	// DeleteDeadBefore deletes up to limit sessions that died before cutoff,
	// revoked before it or with every refresh token expired before it, and
	// their refresh tokens with them. It returns how many sessions it deleted.
	// A session still alive at cutoff, or created after it, is never deleted,
	// and neither is a refresh token on its own.
	DeleteDeadBefore(ctx context.Context, cutoff time.Time, limit int) (int, error)
}

// RefreshToken

type RefreshTokenWriter interface {
	Add(context.Context, *session.RefreshToken) error
	// MarkUsed persists the use of a token already spent in memory. It applies
	// only while the stored token is unused, and returns
	// session.ErrTokenAlreadyUsed when another use got there first.
	MarkUsed(context.Context, *session.RefreshToken) error
}

type RefreshTokenReader interface {
	FindByHash(context.Context, session.RefreshTokenHash) (*session.RefreshToken, error)
}
