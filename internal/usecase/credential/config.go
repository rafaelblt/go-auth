package credential

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	UserReader      port.UserReader
	PasswordReader  port.PasswordReader
	PasswordChecker port.PasswordChecker

	// DummyPasswordHash is verified when the account is missing. It must come
	// from the configured hasher, so it costs the same as a stored hash.
	DummyPasswordHash password.Hashed
}

func NewChecker(cfg Config) (Checker, error) {
	if cfg.UserReader == nil {
		return Checker{}, errors.New("user reader cannot be nil")
	}
	if cfg.PasswordReader == nil {
		return Checker{}, errors.New("password reader cannot be nil")
	}
	if cfg.PasswordChecker == nil {
		return Checker{}, errors.New("password checker cannot be nil")
	}
	if cfg.DummyPasswordHash.IsZero() {
		return Checker{}, errors.New("dummy password hash cannot be zero")
	}
	c := Checker{
		users:      cfg.UserReader,
		passwords:  cfg.PasswordReader,
		pwdChecker: cfg.PasswordChecker,
		dummyHash:  cfg.DummyPasswordHash,
	}
	return c, nil
}
