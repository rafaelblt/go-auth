package port

import (
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
)

type AccessTokenIssuer interface {
	Issue(AccessTokenPayload) (session.AccessToken, error)
}

type AccessTokenValidator interface {
	Validate(raw string) (AccessTokenClaims, error)
}

type AccessTokenPayload struct {
	UserID user.ID
}

type AccessTokenClaims struct {
	UserID user.ID
}
