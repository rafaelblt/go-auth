package usecase

import (
	"context"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Clock interface {
	UtcNow() time.Time
}

type UserReader interface {
	FindByID(context.Context, domain.UserID) (*domain.User, error)
}
type UserExistsChecker interface {
	ExistsByUsername(context.Context, domain.Username) (bool, error)
}
type UserWriter interface {
	Save(context.Context, *domain.User) error
}

type AuthCredentialWriter interface {
	Save(context.Context, *domain.AuthCredential) error
}

type PasswordHasher interface {
	Hash(domain.PlainPassword) (domain.CredentialSecret, error)
	Verify(domain.PlainPassword, domain.CredentialSecret) (bool, error)
}
