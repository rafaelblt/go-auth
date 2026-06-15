package refreshtoken

import (
	"context"
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
)

type Resolver struct {
	reader port.RefreshTokenReader
}

type ResolverConfig struct {
	RefreshTokenReader port.RefreshTokenReader
}

func NewResolver(cfg ResolverConfig) (*Resolver, error) {
	if cfg.RefreshTokenReader == nil {
		return nil, errors.New("refresh token reader nil")
	}
	resolver := Resolver{reader: cfg.RefreshTokenReader}
	return &resolver, nil
}

func (r *Resolver) Resolve(ctx context.Context, raw string) (*session.RefreshToken, error) {
	token, err := decodeToken(raw)
	if err != nil {
		return nil, err
	}
	hash, err := hashToken(token)
	if err != nil {
		return nil, err
	}

	refreshToken, err := r.reader.FindByHash(ctx, hash)
	if err != nil {
		return nil, fmt.Errorf("refresh token reader failed: %w", err)
	}
	if refreshToken == nil {
		return nil, session.ErrTokenInvalid
	}

	return refreshToken, nil
}
