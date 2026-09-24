package port

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type AccessTokenIssuer interface {
	Issue(AccessTokenPayload) (AccessTokenIssued, error)
}

type AccessTokenValidator interface {
	Validate(raw string) (AccessTokenClaims, error)
}

type AccessTokenPayload struct {
	UserID user.ID
}

type AccessTokenIssued struct {
	Token     session.AccessToken
	IssuedAt  time.Time
	ExpiresAt time.Time
}

type AccessTokenClaims struct {
	UserID user.ID
}
