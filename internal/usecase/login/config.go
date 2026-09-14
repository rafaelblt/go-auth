package login

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	UserReader            port.UserReader
	PasswordReader        port.PasswordReader
	PasswordChecker       port.PasswordChecker
	AccessTokenIssuer     port.AccessTokenIssuer
	RefreshTokenGenerator port.RefreshTokenGenerator
	UnitOfWork            port.UnitOfWork
	Clock                 port.Clock

	RefreshTokenTTL time.Duration

	// DummyPasswordHash is verified when the account is missing. It must come
	// from the configured hasher, so it costs the same as a stored hash.
	DummyPasswordHash password.Hashed
}

func New(cfg Config) (Login, error) {
	if cfg.UserReader == nil {
		return Login{}, errors.New("user reader cannot be nil")
	}
	if cfg.PasswordReader == nil {
		return Login{}, errors.New("password reader cannot be nil")
	}
	if cfg.PasswordChecker == nil {
		return Login{}, errors.New("password checker cannot be nil")
	}
	if cfg.AccessTokenIssuer == nil {
		return Login{}, errors.New("access token issuer cannot be nil")
	}
	if cfg.RefreshTokenGenerator == nil {
		return Login{}, errors.New("refresh token generator cannot be nil")
	}
	if cfg.UnitOfWork == nil {
		return Login{}, errors.New("unit of work cannot be nil")
	}
	if cfg.Clock == nil {
		return Login{}, errors.New("clock cannot be nil")
	}
	if cfg.RefreshTokenTTL <= 0 {
		return Login{}, errors.New("refresh ttl zero or negative")
	}
	if cfg.DummyPasswordHash.IsZero() {
		return Login{}, errors.New("dummy password hash cannot be zero")
	}
	uc := Login{
		users:            cfg.UserReader,
		passwords:        cfg.PasswordReader,
		pwdChecker:       cfg.PasswordChecker,
		accessIssuer:     cfg.AccessTokenIssuer,
		refreshGenerator: cfg.RefreshTokenGenerator,
		uow:              cfg.UnitOfWork,
		clock:            cfg.Clock,
		refreshTTL:       cfg.RefreshTokenTTL,
		dummyHash:        cfg.DummyPasswordHash,
	}
	return uc, nil
}
