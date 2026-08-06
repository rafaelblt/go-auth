package bootstrap

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/rafaelblt/go-auth/internal/bootstrap/env"
)

type Config struct {
	Address         string
	DatabaseURL     string
	BcryptCost      int
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func LoadConfig() (Config, error) {
	err := godotenv.Load()
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("godotenv load failed: %w", err)
	}

	cfg := Config{}
	errs := make([]error, 0)

	cfg.DatabaseURL, errs = resolveEnv(EnvDatabaseURL, errs)
	cfg.Address, errs = resolveEnv(EnvAddress, errs)
	cfg.BcryptCost, errs = resolveEnv(EnvBcryptCost, errs)
	cfg.RefreshTokenTTL, errs = resolveEnv(EnvRefreshTokenTTL, errs)
	cfg.AccessTokenTTL, errs = resolveEnv(EnvAccessTokenTTL, errs)

	if len(errs) > 0 {
		e := fmt.Errorf("failed to load environment variables: %w", errors.Join(errs...))
		return Config{}, e
	}

	return cfg, nil
}

func resolveEnv[T any](env env.EnvVar[T], errs []error) (T, []error) {
	value, err := env.Resolve()
	if err != nil {
		e := fmt.Errorf("'%s': %w", env.Key, err)
		errs = append(errs, e)
	}
	return value, errs
}

func (c Config) Validate() error {
	if c.Address == "" {
		return errors.New("address empty")
	}
	if c.DatabaseURL == "" {
		return errors.New("database url empty")
	}
	if c.BcryptCost <= 0 {
		return errors.New("bcrypt cost zero or negative")
	}
	if c.AccessTokenTTL <= 0 {
		return errors.New("access token ttl zero or negative")
	}
	if c.RefreshTokenTTL <= 0 {
		return errors.New("refresh token ttl zero or negative")
	}
	return nil
}
