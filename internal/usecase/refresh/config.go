package refresh

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	SessionReader      port.SessionReader
	AccessTokenIssuer  port.AccessTokenIssuer
	RefreshTokenReader port.RefreshTokenReader
	UnitOfWork         port.UnitOfWork
	Clock              port.Clock

	RefreshTokenTTL time.Duration
}

func New(cfg Config) (*Refresh, error) {
	if cfg.SessionReader == nil {
		return nil, errors.New("session reader nil")
	}
	if cfg.AccessTokenIssuer == nil {
		return nil, errors.New("access token issuer nil")
	}
	if cfg.RefreshTokenReader == nil {
		return nil, errors.New("refresh token reader nil")
	}
	if cfg.UnitOfWork == nil {
		return nil, errors.New("unit of work nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return nil, errors.New("refresh ttl zero or negative")
	}
	uc := Refresh{
		sessions:     cfg.SessionReader,
		accessIssuer: cfg.AccessTokenIssuer,
		tokens:       cfg.RefreshTokenReader,
		uow:          cfg.UnitOfWork,
		clock:        cfg.Clock,
		refreshTTL:   cfg.RefreshTokenTTL,
	}
	return &uc, nil
}
