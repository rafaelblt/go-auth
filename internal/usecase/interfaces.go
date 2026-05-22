package usecase

import (
	"context"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type Clock interface {
	UtcNow() time.Time
}

type PasswordHasher interface {
	Hash(domain.PlainPassword) (domain.CredentialSecret, error)
	Verify(domain.PlainPassword, domain.CredentialSecret) (bool, error)
}

type UnitOfWork interface {
	Do(ctx context.Context, fn func(deps UowDeps) error) error
}
type UowDeps struct {
	UserWriter       UserWriter
	CredentialWriter CredentialWriter
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

type CredentialWriter interface {
	Save(context.Context, *domain.Credential) error
}

type AccessTokenService interface {
	AccessTokenIssuer
	AccessTokenValidator
}

type AccessTokenIssuer interface {
	Issue(AccessTokenPayload) (AccessToken, error)
}

type AccessTokenValidator interface {
	Validate(raw string) (AccessTokenClaims, error)
}

type AccessTokenPayload struct {
	UserID domain.UserID
}

type AccessToken struct {
	Raw       string
	ExpiresAt time.Time
}

type AccessTokenClaims struct {
	UserID domain.UserID
}
