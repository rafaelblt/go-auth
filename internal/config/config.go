package config

import (
	"errors"
	"fmt"
	"time"

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

var envKeyByField = map[string]string{
	fieldAddress:         envAddress.Key,
	fieldDatabaseURL:     envDatabaseURL.Key,
	fieldAutoMigrate:     envAutoMigrate.Key,
	fieldBcryptCost:      envBcryptCost.Key,
	fieldAccessTokenTTL:  envAccessTokenTTL.Key,
	fieldRefreshTokenTTL: envRefreshTokenTTL.Key,
	fieldLogFormat:       envLogFormat.Key,
}

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
	acc := validation.NewAccumulator()

	acc.Add(fieldAddress, validation.Validate(
		params.Address,
		validation.Required[string](),
	))
	acc.Add(fieldDatabaseURL, validation.Validate(
		params.DatabaseURL,
		validation.Required[string](),
	))

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
		autoMigrate:     params.AutoMigrate,
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
		joined := errors.Join(translateToEnvKeys(err)...)
		e := fmt.Errorf("invalid environment variable values: %w", joined)
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

// translateToEnvKeys rewrites the field errors NewConfig reports in terms of
// the environment variable that carried each value. An error of another shape
// is passed through untouched.
func translateToEnvKeys(err error) []error {
	var verr validation.ValidationError
	if !errors.As(err, &verr) {
		return []error{err}
	}

	fieldErrs := verr.Errors()
	errs := make([]error, 0, len(fieldErrs))

	for _, fieldErr := range fieldErrs {
		e := fmt.Errorf("'%s': %s", envKeyOf(fieldErr.Field()), fieldErr.Issue())
		errs = append(errs, e)
	}

	return errs
}

func envKeyOf(field string) string {
	key, ok := envKeyByField[field]
	if !ok {
		return field
	}
	return key
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
