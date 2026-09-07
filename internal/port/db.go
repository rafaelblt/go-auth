package port

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
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
	Add(context.Context, *user.User) error
}

// Password

type PasswordReader interface {
	FindByID(context.Context, password.ID) (*password.Password, error)
	FindByUserID(context.Context, user.ID) (*password.Password, error)
}

type PasswordWriter interface {
	Add(context.Context, *password.Password) error
}

// Session

type SessionReader interface {
	FindByID(context.Context, session.SessionID) (*session.Session, error)
}

type SessionWriter interface {
	Add(context.Context, *session.Session) error
	Update(context.Context, *session.Session) error
}

// RefreshToken

type RefreshTokenWriter interface {
	Add(context.Context, *session.RefreshToken) error
	Update(context.Context, *session.RefreshToken) error
}

type RefreshTokenReader interface {
	FindByHash(context.Context, session.RefreshTokenHash) (*session.RefreshToken, error)
}
