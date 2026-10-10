package login

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
)

type Config struct {
	UserReader        port.UserReader
	PasswordReader    port.PasswordReader
	PasswordChecker   port.PasswordChecker
	AccessTokenIssuer port.AccessTokenIssuer
	UnitOfWork        port.UnitOfWork
	Clock             port.Clock

	RefreshTokenTTL time.Duration

	// DummyPasswordHash is verified when the account is missing. It must come
	// from the configured hasher, so it costs the same as a stored hash.
	DummyPasswordHash password.Hashed
}

func New(cfg Config) (Login, error) {
	credentials, err := credential.NewChecker(credential.Config{
		UserReader:        cfg.UserReader,
		PasswordReader:    cfg.PasswordReader,
		PasswordChecker:   cfg.PasswordChecker,
		DummyPasswordHash: cfg.DummyPasswordHash,
	})
	if err != nil {
		return Login{}, err
	}
	if cfg.AccessTokenIssuer == nil {
		return Login{}, errors.New("access token issuer cannot be nil")
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
	uc := Login{
		credentials:  credentials,
		accessIssuer: cfg.AccessTokenIssuer,
		uow:          cfg.UnitOfWork,
		clock:        cfg.Clock,
		refreshTTL:   cfg.RefreshTokenTTL,
	}
	return uc, nil
}
