package changepassword

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase/credential"
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
	credentials, err := credential.NewChecker(credential.Config{
		UserReader:        cfg.UserReader,
		PasswordReader:    cfg.PasswordReader,
		PasswordChecker:   cfg.PasswordChecker,
		DummyPasswordHash: cfg.DummyPasswordHash,
	})
	if err != nil {
		return ChangePassword{}, err
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
	uc := ChangePassword{
		credentials: credentials,
		hasher:      cfg.PasswordHasher,
		uow:         cfg.UnitOfWork,
		clock:       cfg.Clock,
	}
	return uc, nil
}
