package usecase

import (
	"context"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Clock interface {
	UtcNow() time.Time
}

type UserFinder interface {
	FindByID(context.Context, domain.UserID) (*domain.User, error)
}
type UserExistsChecker interface {
	ExistsByUsername(context.Context, domain.Username) (bool, error)
}
type UserSaver interface {
	Save(context.Context, *domain.User) error
}

type UserCredentialsSaver interface {
	Save(context.Context, *domain.UserCredentials) error
}

type PasswordHasher interface {
	Hash(domain.PlainPassword) (domain.HashedPassword, error)
	Verify(domain.PlainPassword, domain.HashedPassword) (bool, error)
}
