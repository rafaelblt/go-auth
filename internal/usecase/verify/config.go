package verify

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	AccessTokenValidator port.AccessTokenValidator
	Clock                port.Clock
}

func New(cfg Config) (*Verify, error) {
	if cfg.AccessTokenValidator == nil {
		return nil, errors.New("access token validator nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	uc := Verify{
		validator: cfg.AccessTokenValidator,
		clock:     cfg.Clock,
	}
	return &uc, nil
}
