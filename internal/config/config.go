package config

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	address         string
	databaseURL     string
	autoMigrate     bool
	bcryptCost      int
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logFormat       LogFormat
}

type ConfigParams struct {
	Address         string
	DatabaseURL     string
	AutoMigrate     bool
	BcryptCost      *int
	AccessTokenTTL  *time.Duration
	RefreshTokenTTL *time.Duration
	LogFormat       *LogFormat
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

	logFormat := defaultLogFormat
	if params.LogFormat != nil {
		if !params.LogFormat.valid() {
			errs = append(errs, fmt.Errorf("log format %q unknown", *params.LogFormat))
		}
		logFormat = *params.LogFormat
	}

	if len(errs) > 0 {
		e := fmt.Errorf("invalid config params: %w", errors.Join(errs...))
		return Config{}, e
	}

	cfg := Config{
		address:         params.Address,
		databaseURL:     params.DatabaseURL,
		autoMigrate:     params.AutoMigrate,
		bcryptCost:      bcryptCost,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		logFormat:       logFormat,
	}
	return cfg, nil
}

func LoadConfig() (Config, error) {
	errs := make([]error, 0)

	databaseURL, errs := resolveEnv(envDatabaseURL, errs)
	address, errs := resolveEnv(envAddress, errs)
	autoMigrate, errs := resolveEnv(envAutoMigrate, errs)
	bcryptCost, errs := resolveEnv(envBcryptCost, errs)
	refreshTokenTTL, errs := resolveEnv(envRefreshTokenTTL, errs)
	accessTokenTTL, errs := resolveEnv(envAccessTokenTTL, errs)
	logFormat, errs := resolveEnv(envLogFormat, errs)

	if len(errs) > 0 {
		e := fmt.Errorf("failed to load environment variables: %w", errors.Join(errs...))
		return Config{}, e
	}

	cfg, err := NewConfig(ConfigParams{
		Address:         address,
		DatabaseURL:     databaseURL,
		AutoMigrate:     autoMigrate,
		BcryptCost:      &bcryptCost,
		AccessTokenTTL:  &accessTokenTTL,
		RefreshTokenTTL: &refreshTokenTTL,
		LogFormat:       &logFormat,
	})
	if err != nil {
		e := fmt.Errorf("configuration from environment variables invalid: %w", err)
		return Config{}, e
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

func (c Config) IsZero() bool {
	return c == Config{}
}

func (c Config) Address() string {
	return c.address
}

func (c Config) DatabaseURL() string {
	return c.databaseURL
}

func (c Config) AutoMigrate() bool {
	return c.autoMigrate
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

func (c Config) LogFormat() LogFormat {
	return c.logFormat
}
