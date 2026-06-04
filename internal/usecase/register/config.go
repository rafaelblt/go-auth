package register

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	UserExistsChecker port.UserExistsChecker
	UnitOfWork        port.UnitOfWork
	PasswordHasher    port.PasswordHasher
	Clock             port.Clock
}

func New(cfg Config) (Register, error) {
	if cfg.UserExistsChecker == nil {
		return Register{}, errors.New("user exists checker cannot be nil")
	}
	if cfg.UnitOfWork == nil {
		return Register{}, errors.New("unit of work cannot be nil")
	}
	if cfg.PasswordHasher == nil {
		return Register{}, errors.New("password hasher cannot be nil")
	}
	if cfg.Clock == nil {
		return Register{}, errors.New("clock cannot be nil")
	}
	uc := Register{
		userExists: cfg.UserExistsChecker,
		uow:        cfg.UnitOfWork,
		hasher:     cfg.PasswordHasher,
		clock:      cfg.Clock,
	}
	return uc, nil
}
