package config

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	address         string
	databaseURL     string
	bcryptCost      int
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
}

type ConfigParams struct {
	Address         string
	DatabaseURL     string
	BcryptCost      *int
	AccessTokenTTL  *time.Duration
	RefreshTokenTTL *time.Duration
}

func NewConfig(params ConfigParams) (Config, error) {
	errs := make([]error, 0)

	if params.Address == "" {
		errs = append(errs, errors.New("address empty"))
	}
	if params.DatabaseURL == "" {
		errs = append(errs, errors.New("database url empty"))
	}

	bcryptCost := defaultBcryptCost
	if params.BcryptCost != nil {
		if *params.BcryptCost <= 0 {
			errs = append(errs, errors.New("bcrypt cost zero or negative"))
		}
		bcryptCost = *params.BcryptCost
	}

	accessTokenTTL := defaultAccessTokenTTL
	if params.AccessTokenTTL != nil {
		if *params.AccessTokenTTL <= 0 {
			errs = append(errs, errors.New("access token ttl zero or negative"))
		}
		accessTokenTTL = *params.AccessTokenTTL
	}

	refreshTokenTTL := defaultRefreshTokenTTL
	if params.RefreshTokenTTL != nil {
		if *params.RefreshTokenTTL <= 0 {
			errs = append(errs, errors.New("refresh token ttl zero or negative"))
		}
		refreshTokenTTL = *params.RefreshTokenTTL
	}

	if len(errs) > 0 {
		e := fmt.Errorf("invalid config params: %w", errors.Join(errs...))
		return Config{}, e
	}

	cfg := Config{
		address:         params.Address,
		databaseURL:     params.DatabaseURL,
		bcryptCost:      bcryptCost,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
	}
	return cfg, nil
}

func LoadConfig() (Config, error) {
	errs := make([]error, 0)

	databaseURL, errs := resolveEnv(envDatabaseURL, errs)
	address, errs := resolveEnv(envAddress, errs)
	bcryptCost, errs := resolveEnv(envBcryptCost, errs)
	refreshTokenTTL, errs := resolveEnv(envRefreshTokenTTL, errs)
	accessTokenTTL, errs := resolveEnv(envAccessTokenTTL, errs)

	if len(errs) > 0 {
		e := fmt.Errorf("failed to load environment variables: %w", errors.Join(errs...))
		return Config{}, e
	}

	cfg, err := NewConfig(ConfigParams{
		Address:         address,
		DatabaseURL:     databaseURL,
		BcryptCost:      &bcryptCost,
		AccessTokenTTL:  &accessTokenTTL,
		RefreshTokenTTL: &refreshTokenTTL,
	})
	if err != nil {
		panic(fmt.Sprintf("config loaded from environment variables invalid for new config: %s", err))
	}

	return cfg, nil
}

func resolveEnv[T any](ev env[T], errs []error) (T, []error) {
	value, err := ev.Resolve()
	if err != nil {
		e := fmt.Errorf("'%s': %w", ev.Key, err)
		errs = append(errs, e)
	}
	return value, errs
}

// IsZero reports whether the Config was never built, a Config coming from
// NewConfig or LoadConfig always has every field set.
func (c Config) IsZero() bool {
	return c == Config{}
}

func (c Config) Address() string {
	return c.address
}

func (c Config) DatabaseURL() string {
	return c.databaseURL
}

func (c Config) BcryptCost() int {
	return c.bcryptCost
}

func (c Config) AccessTokenTTL() time.Duration {
	return c.accessTokenTTL
}

func (c Config) RefreshTokenTTL() time.Duration {
	return c.refreshTokenTTL
}
