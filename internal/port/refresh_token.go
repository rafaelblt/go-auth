package port

import (
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
)

type RefreshTokenIssuer interface {
	Issue(RefreshTokenPayload) (RefreshTokenIssued, error)
}

type RefreshTokenRotator interface {
	Rotate(raw string) (RefreshTokenRotated, error)
}

type RefreshTokenPayload struct {
	UserID    user.ID
	SessionID session.SessionID
}

type RefreshTokenIssued struct {
	Token    *session.RefreshToken
	RawValue string
}

type RefreshTokenRotated struct {
	Used   session.RefreshToken
	Issued RefreshTokenIssued
}
