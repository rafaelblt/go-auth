package port

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
)

// Unit Of Work

type UnitOfWork interface {
	Do(ctx context.Context, fn func(deps UowDeps) error) error
}

type UowDeps struct {
	UserWriter         UserWriter
	CredentialWriter   CredentialWriter
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

// Credential

type CredentialReader interface {
	FindByID(context.Context, credential.ID) (*credential.Credential, error)
	FindByUserAndKind(context.Context, user.ID, credential.Kind) (*credential.Credential, error)
}

type CredentialWriter interface {
	Add(context.Context, *credential.Credential) error
}

// Session

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
