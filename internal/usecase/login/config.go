package login

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	UserReader         port.UserReader
	CredentialReader   port.CredentialReader
	PasswordChecker    port.PasswordChecker
	AccessTokenIssuer  port.AccessTokenIssuer
	RefreshTokenIssuer port.RefreshTokenIssuer
	UnitOfWork         port.UnitOfWork
	Clock              port.Clock
}

func New(cfg Config) (Login, error) {
	if cfg.UserReader == nil {
		return Login{}, errors.New("user reader cannot be nil")
	}
	if cfg.CredentialReader == nil {
		return Login{}, errors.New("credential reader cannot be nil")
	}
	if cfg.PasswordChecker == nil {
		return Login{}, errors.New("password checker cannot be nil")
	}
	if cfg.AccessTokenIssuer == nil {
		return Login{}, errors.New("access token issuer cannot be nil")
	}
	if cfg.RefreshTokenIssuer == nil {
		return Login{}, errors.New("refresh token issuer cannot be nil")
	}
	if cfg.UnitOfWork == nil {
		return Login{}, errors.New("unit of work cannot be nil")
	}
	if cfg.Clock == nil {
		return Login{}, errors.New("clock cannot be nil")
	}
	uc := Login{
		users:         cfg.UserReader,
		credentials:   cfg.CredentialReader,
		pwdChecker:    cfg.PasswordChecker,
		accessIssuer:  cfg.AccessTokenIssuer,
		refreshIssuer: cfg.RefreshTokenIssuer,
		uow:           cfg.UnitOfWork,
		clock:         cfg.Clock,
	}
	return uc, nil
}
