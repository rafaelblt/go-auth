package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/validation"
)

// Fields tagged on the errors NewConfig returns. LoadConfig maps them back to
// the environment variable that carried the value, so a user configuring the
// app through the environment reads the key they actually set.
const (
	fieldAddress         = "address"
	fieldDatabaseURL     = "database_url"
	fieldAutoMigrate     = "auto_migrate"
	fieldBcryptCost      = "bcrypt_cost"
	fieldAccessTokenTTL  = "access_token_ttl"
	fieldRefreshTokenTTL = "refresh_token_ttl"
	fieldLogFormat       = "log_format"
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
	AutoMigrate     *bool
	BcryptCost      *int
	AccessTokenTTL  *time.Duration
	RefreshTokenTTL *time.Duration
	LogFormat       *LogFormat
}

func NewConfig(params ConfigParams) (Config, error) {
	acc := validation.NewAccumulator()

	acc.Add(fieldAddress, validation.Validate(
		params.Address,
		validation.Required[string](),
	))
	acc.Add(fieldDatabaseURL, validation.Validate(
		params.DatabaseURL,
		validation.Required[string](),
	))

	autoMigrate := resolveParam(acc, fieldAutoMigrate, params.AutoMigrate,
		defaultAutoMigrate)
	bcryptCost := resolveParam(acc, fieldBcryptCost, params.BcryptCost,
		defaultBcryptCost, validation.Positive[int]())
	accessTokenTTL := resolveParam(acc, fieldAccessTokenTTL, params.AccessTokenTTL,
		defaultAccessTokenTTL, validation.Positive[time.Duration]())
	refreshTokenTTL := resolveParam(acc, fieldRefreshTokenTTL, params.RefreshTokenTTL,
		defaultRefreshTokenTTL, validation.Positive[time.Duration]())
	logFormat := resolveParam(acc, fieldLogFormat, params.LogFormat,
		defaultLogFormat, allowedLogFormat())

	if err := acc.Err(); err != nil {
		return Config{}, err
	}

	cfg := Config{
		address:         params.Address,
		databaseURL:     params.DatabaseURL,
		autoMigrate:     autoMigrate,
		bcryptCost:      bcryptCost,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		logFormat:       logFormat,
	}
	return cfg, nil
}

// resolveParam validates an optional param when it was provided, and falls
// back to the default when it was not.
func resolveParam[T any](
	acc *validation.Accumulator,
	field string,
	param *T,
	fallback T,
	validators ...validation.Validator[T],
) T {
	if param == nil {
		return fallback
	}

	acc.Add(field, validation.Validate(*param, validators...))
	return *param
}

func LoadConfig() (Config, error) {
	load := newEnvLoad()

	address := resolveEnv(load, envAddress, fieldAddress)
	databaseURL := resolveEnv(load, envDatabaseURL, fieldDatabaseURL)
	autoMigrate := resolveEnv(load, envAutoMigrate, fieldAutoMigrate)
	bcryptCost := resolveEnv(load, envBcryptCost, fieldBcryptCost)
	accessTokenTTL := resolveEnv(load, envAccessTokenTTL, fieldAccessTokenTTL)
	refreshTokenTTL := resolveEnv(load, envRefreshTokenTTL, fieldRefreshTokenTTL)
	logFormat := resolveEnv(load, envLogFormat, fieldLogFormat)

	cfg, err := NewConfig(ConfigParams{
		Address:         shared.Deref(address),
		DatabaseURL:     shared.Deref(databaseURL),
		AutoMigrate:     autoMigrate,
		BcryptCost:      bcryptCost,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		LogFormat:       logFormat,
	})

	errs := load.merge(err)
	if len(errs) > 0 {
		e := fmt.Errorf("invalid environment configuration: %w", errors.Join(errs...))
		return Config{}, e
	}

	return cfg, nil
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
