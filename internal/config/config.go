// Package config loads and checks the service's settings, from environment
// variables or directly from code. Each setting is checked against the rules
// of the code its value is passed to, and every problem is reported at once,
// before anything is built, so the service never starts half-configured.
//
// See docs/architecture/startup.md#configuration-loading.
package config

import (
	"net/netip"
	"reflect"
	"slices"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/infra/bcrypt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt"
	"github.com/rafaelblt/go-auth/internal/infra/jwt/ed25519"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
)

type Config struct {
	address         string
	databaseURL     string
	autoMigrate     bool
	bcryptCost      int
	accessTokenTTL  time.Duration
	refreshTokenTTL time.Duration
	logFormat       LogFormat
	rateLimit       RateLimitLevel
	trustedProxies  []netip.Prefix

	signingKeyEncryptionKey []byte
	sessionRetention        SessionRetention
}

type ConfigParams struct {
	Address         string
	DatabaseURL     string
	AutoMigrate     *bool
	BcryptCost      *int
	AccessTokenTTL  *time.Duration
	RefreshTokenTTL *time.Duration
	LogFormat       *LogFormat
	RateLimit       *RateLimitLevel
	TrustedProxies  []netip.Prefix
	// SigningKeyEncryptionKey is the raw 32-byte key. Nil or empty means not
	// set.
	SigningKeyEncryptionKey []byte
	SessionRetention        *SessionRetention
}

// Fallbacks for the params NewConfig accepts as optional.
const (
	defaultAutoMigrate     = false
	defaultBcryptCost      = 12
	defaultAccessTokenTTL  = 30 * time.Minute
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
	defaultLogFormat       = LogFormatJSON
	defaultRateLimit       = RateLimitOff

	defaultSessionRetention = 30 * 24 * time.Hour
)

func NewConfig(params ConfigParams) (Config, error) {
	cfg, problems := newConfig(params)
	if len(problems) > 0 {
		return Config{}, newConfigError(problems)
	}
	return cfg, nil
}

// newConfig applies the defaults and checks every setting, with the rules the
// code its value is passed to exports rather than a copy of them.
//
// See docs/architecture/startup.md#configuration-loading.
func newConfig(params ConfigParams) (Config, []problem) {
	autoMigrate := shared.DerefOr(params.AutoMigrate, defaultAutoMigrate)
	bcryptCost := shared.DerefOr(params.BcryptCost, defaultBcryptCost)
	accessTokenTTL := shared.DerefOr(params.AccessTokenTTL, defaultAccessTokenTTL)
	refreshTokenTTL := shared.DerefOr(params.RefreshTokenTTL, defaultRefreshTokenTTL)
	logFormat := shared.DerefOr(params.LogFormat, defaultLogFormat)
	rateLimit := shared.DerefOr(params.RateLimit, defaultRateLimit)
	sessionRetention := shared.DerefOr(params.SessionRetention, SessionRetentionOf(defaultSessionRetention))

	var problems []problem
	check(&problems, envAddress.Key, params.Address, required, hostPort)
	check(&problems, envDatabaseURL.Key, params.DatabaseURL, required, postgres.CheckConnString)
	check(&problems, envBcryptCost.Key, bcryptCost, bcrypt.CheckCost)
	check(&problems, envAccessTokenTTL.Key, accessTokenTTL, jwt.CheckExpiration)
	check(&problems, envRefreshTokenTTL.Key, refreshTokenTTL, session.CheckRefreshTokenTTL)
	check(&problems, envLogFormat.Key, logFormat, checkLogFormat)
	check(&problems, envRateLimit.Key, rateLimit, checkRateLimitLevel)
	check(&problems, envSigningKeyEncryptionKey.Key, params.SigningKeyEncryptionKey, ed25519.CheckEncryptionKey)
	if retention, on := sessionRetention.Duration(); on {
		check(&problems, envSessionRetention.Key, retention, sessioncleanup.CheckRetention)
	}

	if len(problems) > 0 {
		return Config{}, problems
	}

	var signingKeyEncryptionKey []byte
	if len(params.SigningKeyEncryptionKey) > 0 {
		signingKeyEncryptionKey = slices.Clone(params.SigningKeyEncryptionKey)
	}

	cfg := Config{
		address:         params.Address,
		databaseURL:     params.DatabaseURL,
		autoMigrate:     autoMigrate,
		bcryptCost:      bcryptCost,
		accessTokenTTL:  accessTokenTTL,
		refreshTokenTTL: refreshTokenTTL,
		logFormat:       logFormat,
		rateLimit:       rateLimit,
		trustedProxies:  slices.Clone(params.TrustedProxies),

		signingKeyEncryptionKey: signingKeyEncryptionKey,
		sessionRetention:        sessionRetention,
	}
	return cfg, nil
}

func LoadConfig() (Config, error) {
	load := newEnvLoad()

	address := resolveEnv(load, envAddress)
	databaseURL := resolveEnv(load, envDatabaseURL)
	autoMigrate := resolveEnv(load, envAutoMigrate)
	bcryptCost := resolveEnv(load, envBcryptCost)
	accessTokenTTL := resolveEnv(load, envAccessTokenTTL)
	refreshTokenTTL := resolveEnv(load, envRefreshTokenTTL)
	logFormat := resolveEnv(load, envLogFormat)
	rateLimit := resolveEnv(load, envRateLimit)
	trustedProxies := resolveEnv(load, envTrustedProxies)
	signingKeyEncryptionKey := resolveEnv(load, envSigningKeyEncryptionKey)
	sessionRetention := resolveEnv(load, envSessionRetention)

	cfg, problems := newConfig(ConfigParams{
		Address:         shared.Deref(address),
		DatabaseURL:     shared.Deref(databaseURL),
		AutoMigrate:     autoMigrate,
		BcryptCost:      bcryptCost,
		AccessTokenTTL:  accessTokenTTL,
		RefreshTokenTTL: refreshTokenTTL,
		LogFormat:       logFormat,
		RateLimit:       rateLimit,
		TrustedProxies:  shared.Deref(trustedProxies),

		SigningKeyEncryptionKey: shared.Deref(signingKeyEncryptionKey),
		SessionRetention:        sessionRetention,
	})

	problems = load.merge(problems)
	if len(problems) > 0 {
		return Config{}, newConfigError(problems)
	}

	return cfg, nil
}

func (c Config) IsZero() bool {
	return reflect.DeepEqual(c, Config{})
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

func (c Config) RateLimit() RateLimitLevel {
	return c.rateLimit
}

func (c Config) TrustedProxies() []netip.Prefix {
	return slices.Clone(c.trustedProxies)
}

func (c Config) SigningKeyEncryptionKey() []byte {
	return slices.Clone(c.signingKeyEncryptionKey)
}

func (c Config) SessionRetention() SessionRetention {
	return c.sessionRetention
}
