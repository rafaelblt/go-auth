package port

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/domain/session"
)

type RefreshTokenGenerator interface {
	Generate() (RefreshTokenGenerated, error)
}

type RefreshTokenResolver interface {
	Resolve(ctx context.Context, raw string) (*session.RefreshToken, error)
}

type RefreshTokenGenerated struct {
	Raw  string
	Hash session.RefreshTokenHash
}
