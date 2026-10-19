package logout

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	SessionReader      port.SessionReader
	RefreshTokenReader port.RefreshTokenReader
	UnitOfWork         port.UnitOfWork
	Clock              port.Clock
}

func New(cfg Config) (*Logout, error) {
	if cfg.SessionReader == nil {
		return nil, errors.New("session reader nil")
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
	uc := Logout{
		sessions: cfg.SessionReader,
		tokens:   cfg.RefreshTokenReader,
		uow:      cfg.UnitOfWork,
		clock:    cfg.Clock,
	}
	return &uc, nil
}
