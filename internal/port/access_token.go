package port

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
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
	ExpiresAt time.Time
}

type AccessTokenClaims struct {
	UserID user.ID
}
