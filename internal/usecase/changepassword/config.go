package changepassword

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	UserReader      port.UserReader
	PasswordReader  port.PasswordReader
	PasswordChecker port.PasswordChecker
	PasswordHasher  port.PasswordHasher
	UnitOfWork      port.UnitOfWork
	Clock           port.Clock

	// DummyPasswordHash is verified when the account is missing. It must come
	// from the configured hasher, so it costs the same as a stored hash.
	DummyPasswordHash password.Hashed
}

func New(cfg Config) (ChangePassword, error) {
	if cfg.UserReader == nil {
		return ChangePassword{}, errors.New("user reader cannot be nil")
	}
	if cfg.PasswordReader == nil {
		return ChangePassword{}, errors.New("password reader cannot be nil")
	}
	if cfg.PasswordChecker == nil {
		return ChangePassword{}, errors.New("password checker cannot be nil")
	}
	if cfg.PasswordHasher == nil {
		return ChangePassword{}, errors.New("password hasher cannot be nil")
	}
	if cfg.UnitOfWork == nil {
		return ChangePassword{}, errors.New("unit of work cannot be nil")
	}
	if cfg.Clock == nil {
		return ChangePassword{}, errors.New("clock cannot be nil")
	}
	if cfg.DummyPasswordHash.IsZero() {
		return ChangePassword{}, errors.New("dummy password hash cannot be zero")
	}
	uc := ChangePassword{
		users:      cfg.UserReader,
		passwords:  cfg.PasswordReader,
		pwdChecker: cfg.PasswordChecker,
		hasher:     cfg.PasswordHasher,
		uow:        cfg.UnitOfWork,
		clock:      cfg.Clock,
		dummyHash:  cfg.DummyPasswordHash,
	}
	return uc, nil
}
