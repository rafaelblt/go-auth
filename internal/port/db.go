package port

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/user"
)

// Unit Of Work

type UnitOfWork interface {
	Do(ctx context.Context, fn func(deps UowDeps) error) error
}

type UowDeps struct {
	UserWriter       UserWriter
	CredentialWriter CredentialWriter
}

// User

type UserReader interface {
	FindByID(context.Context, user.ID) (*user.User, error)
}

type UserExistsChecker interface {
	ExistsByUsername(context.Context, user.Username) (bool, error)
}

type UserWriter interface {
	Save(context.Context, *user.User) error
}

// Credential

type CredentialWriter interface {
	Save(context.Context, *credential.Credential) error
}
