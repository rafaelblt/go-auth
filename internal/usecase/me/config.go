package me

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/port"
)

type Config struct {
	AccessTokenValidator port.AccessTokenValidator
	UserReader           port.UserReader
}

func New(cfg Config) (*Me, error) {
	if cfg.AccessTokenValidator == nil {
		return nil, errors.New("access token validator nil")
	}
	if cfg.UserReader == nil {
		return nil, errors.New("user reader nil")
	}
	uc := Me{
		validator: cfg.AccessTokenValidator,
		users:     cfg.UserReader,
	}
	return &uc, nil
}
