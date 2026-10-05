package config

import (
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ReturnsConfig(t *testing.T) {
	address := "add ress vlaue"
	databaseURL := "db url vlaue"
	autoMigrate := true
	bcryptCost := 8
	accessTokenTTL := 22 * time.Minute
	refreshTokenTTL := 2 * 24 * time.Hour
	logFormat := LogFormatText

	t.Setenv(envAddress.Key, address)
	t.Setenv(envDatabaseURL.Key, databaseURL)
	t.Setenv(envAutoMigrate.Key, strconv.FormatBool(autoMigrate))
	t.Setenv(envBcryptCost.Key, strconv.Itoa(bcryptCost))
	t.Setenv(envAccessTokenTTL.Key, accessTokenTTL.String())
	t.Setenv(envRefreshTokenTTL.Key, refreshTokenTTL.String())
	t.Setenv(envLogFormat.Key, string(logFormat))
	t.Setenv(envRateLimit.Key, "STRICT")
	t.Setenv(envTrustedProxies.Key, " 10.0.0.0/8, 192.0.2.1 ")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, address, cfg.Address())
	assert.Equal(t, databaseURL, cfg.DatabaseURL())
	assert.Equal(t, autoMigrate, cfg.AutoMigrate())
	assert.Equal(t, bcryptCost, cfg.BcryptCost())
	assert.Equal(t, accessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, refreshTokenTTL, cfg.RefreshTokenTTL())
	assert.Equal(t, logFormat, cfg.LogFormat())
	assert.Equal(t, RateLimitStrict, cfg.RateLimit())
	assert.Equal(t, []netip.Prefix{
		netip.MustParsePrefix("10.0.0.0/8"),
		netip.MustParsePrefix("192.0.2.1/32"),
	}, cfg.TrustedProxies())
}

func TestLoadConfig_ReturnsDefaults_WhenOptionalEnvsAreMissing(t *testing.T) {
	t.Setenv(envAddress.Key, "add ress vlaue")
	t.Setenv(envDatabaseURL.Key, "db url vlaue")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, defaultAutoMigrate, cfg.AutoMigrate())
	assert.Equal(t, defaultBcryptCost, cfg.BcryptCost())
	assert.Equal(t, defaultAccessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, defaultRefreshTokenTTL, cfg.RefreshTokenTTL())
	assert.Equal(t, defaultLogFormat, cfg.LogFormat())
	assert.Equal(t, RateLimitOff, cfg.RateLimit())
	assert.Empty(t, cfg.TrustedProxies())
}

func setRequiredEnvs(t *testing.T) {
	t.Helper()
	t.Setenv(envAddress.Key, "add ress vlaue")
	t.Setenv(envDatabaseURL.Key, "db url vlaue")
}

func TestLoadConfig_ReturnsError(t *testing.T) {
	testCases := []struct {
		desc  string
		setup func(t *testing.T)
	}{
		{
			desc:  "address empty",
			setup: func(t *testing.T) { t.Setenv(envAddress.Key, "") },
		},
		{
			desc:  "database url empty",
			setup: func(t *testing.T) { t.Setenv(envDatabaseURL.Key, "") },
		},
		{
			desc:  "auto migrate not bool",
			setup: func(t *testing.T) { t.Setenv(envAutoMigrate.Key, "yes") },
		},
		{
			desc:  "bcrypt cost not int",
			setup: func(t *testing.T) { t.Setenv(envBcryptCost.Key, "twelve") },
		},
		{
			desc:  "access token ttl not duration format",
			setup: func(t *testing.T) { t.Setenv(envAccessTokenTTL.Key, "1 hour") },
		},
		{
			desc:  "refresh token ttl not duration format",
			setup: func(t *testing.T) { t.Setenv(envRefreshTokenTTL.Key, "1 month") },
		},
		{
			desc:  "log format unknown",
			setup: func(t *testing.T) { t.Setenv(envLogFormat.Key, "xml") },
		},
		{
			desc: "rate limit unknown",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envRateLimit.Key, "extreme")
			},
		},
		{
			desc: "trusted proxies not an address",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envTrustedProxies.Key, "proxy.local")
			},
		},
		{
			desc: "bcrypt cost zero",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envBcryptCost.Key, "0")
			},
		},
		{
			desc: "bcrypt cost negative",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envBcryptCost.Key, "-1")
			},
		},
		{
			desc: "access token ttl zero",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envAccessTokenTTL.Key, "0s")
			},
		},
		{
			desc: "access token ttl negative",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envAccessTokenTTL.Key, "-5m")
			},
		},
		{
			desc: "refresh token ttl zero",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envRefreshTokenTTL.Key, "0s")
			},
		},
		{
			desc: "refresh token ttl negative",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envRefreshTokenTTL.Key, "-1h")
			},
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			tC.setup(t)

			cfg, err := LoadConfig()

			assert.Error(t, err)
			assert.Zero(t, cfg)
		})
	}
}

func TestNewConfig_ReturnsConfig(t *testing.T) {
	params := ConfigParams{
		Address:         "add ress vlaue",
		DatabaseURL:     "db url vlaue",
		AutoMigrate:     shared.Ptr(true),
		BcryptCost:      shared.Ptr(8),
		AccessTokenTTL:  shared.Ptr(22 * time.Minute),
		RefreshTokenTTL: shared.Ptr(2 * 24 * time.Hour),
		LogFormat:       shared.Ptr(LogFormatText),
		RateLimit:       shared.Ptr(RateLimitNormal),
		TrustedProxies:  []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},
	}

	cfg, err := NewConfig(params)

	require.NoError(t, err)
	assert.False(t, cfg.IsZero())
	assert.Equal(t, params.Address, cfg.Address())
	assert.Equal(t, params.DatabaseURL, cfg.DatabaseURL())
	assert.Equal(t, *params.AutoMigrate, cfg.AutoMigrate())
	assert.Equal(t, *params.BcryptCost, cfg.BcryptCost())
	assert.Equal(t, *params.AccessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, *params.RefreshTokenTTL, cfg.RefreshTokenTTL())
	assert.Equal(t, *params.LogFormat, cfg.LogFormat())
	assert.Equal(t, *params.RateLimit, cfg.RateLimit())
	assert.Equal(t, params.TrustedProxies, cfg.TrustedProxies())
}

func TestNewConfig_ReturnsDefaults_WhenOptionalParamsAreMissing(t *testing.T) {
	params := ConfigParams{
		Address:     "add ress vlaue",
		DatabaseURL: "db url vlaue",
	}

	cfg, err := NewConfig(params)

	require.NoError(t, err)
	assert.Equal(t, defaultAutoMigrate, cfg.AutoMigrate())
	assert.Equal(t, defaultBcryptCost, cfg.BcryptCost())
	assert.Equal(t, defaultAccessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, defaultRefreshTokenTTL, cfg.RefreshTokenTTL())
	assert.Equal(t, defaultLogFormat, cfg.LogFormat())
	assert.Equal(t, RateLimitOff, cfg.RateLimit())
	assert.Empty(t, cfg.TrustedProxies())
}

func TestNewConfig_ReturnsError(t *testing.T) {
	valid := func() ConfigParams {
		return ConfigParams{
			Address:         "add ress vlaue",
			DatabaseURL:     "db url vlaue",
			BcryptCost:      shared.Ptr(8),
			AccessTokenTTL:  shared.Ptr(22 * time.Minute),
			RefreshTokenTTL: shared.Ptr(2 * 24 * time.Hour),
			LogFormat:       shared.Ptr(LogFormatJSON),
		}
	}
	testCases := []struct {
		desc   string
		mutate func(p *ConfigParams)
	}{
		{
			desc:   "address empty",
			mutate: func(p *ConfigParams) { p.Address = "" },
		},
		{
			desc:   "database url empty",
			mutate: func(p *ConfigParams) { p.DatabaseURL = "" },
		},
		{
			desc:   "bcrypt cost zero",
			mutate: func(p *ConfigParams) { p.BcryptCost = shared.Ptr(0) },
		},
		{
			desc:   "bcrypt cost negative",
			mutate: func(p *ConfigParams) { p.BcryptCost = shared.Ptr(-1) },
		},
		{
			desc:   "access token ttl zero",
			mutate: func(p *ConfigParams) { p.AccessTokenTTL = shared.Ptr(time.Duration(0)) },
		},
		{
			desc:   "access token ttl negative",
			mutate: func(p *ConfigParams) { p.AccessTokenTTL = shared.Ptr(-time.Minute) },
		},
		{
			desc:   "refresh token ttl zero",
			mutate: func(p *ConfigParams) { p.RefreshTokenTTL = shared.Ptr(time.Duration(0)) },
		},
		{
			desc:   "refresh token ttl negative",
			mutate: func(p *ConfigParams) { p.RefreshTokenTTL = shared.Ptr(-time.Hour) },
		},
		{
			desc:   "log format empty",
			mutate: func(p *ConfigParams) { p.LogFormat = shared.Ptr(LogFormat("")) },
		},
		{
			desc:   "log format unknown",
			mutate: func(p *ConfigParams) { p.LogFormat = shared.Ptr(LogFormat("xml")) },
		},
		{
			desc:   "rate limit empty",
			mutate: func(p *ConfigParams) { p.RateLimit = shared.Ptr(RateLimitLevel("")) },
		},
		{
			desc:   "rate limit unknown",
			mutate: func(p *ConfigParams) { p.RateLimit = shared.Ptr(RateLimitLevel("extreme")) },
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			params := valid()
			tC.mutate(&params)

			cfg, err := NewConfig(params)

			assert.Error(t, err)
			assert.True(t, cfg.IsZero())
		})
	}
}

func TestConfig_IsZero(t *testing.T) {
	assert.True(t, Config{}.IsZero())
}

func TestConfig_TrustedProxies_IsACopy(t *testing.T) {
	trusted := netip.MustParsePrefix("10.0.0.0/8")
	params := ConfigParams{
		Address:        "add ress vlaue",
		DatabaseURL:    "db url vlaue",
		TrustedProxies: []netip.Prefix{trusted},
	}
	cfg, err := NewConfig(params)
	require.NoError(t, err)

	params.TrustedProxies[0] = netip.MustParsePrefix("0.0.0.0/0")
	cfg.TrustedProxies()[0] = netip.MustParsePrefix("0.0.0.0/0")

	assert.Equal(t, []netip.Prefix{trusted}, cfg.TrustedProxies())
}

func TestNewConfig_ReturnsErrorsTaggedWithTheirField(t *testing.T) {
	params := ConfigParams{BcryptCost: shared.Ptr(0)}

	_, err := NewConfig(params)

	var verr validation.ValidationError
	require.ErrorAs(t, err, &verr)

	codes := make(map[string]string)
	for _, fieldErr := range verr.Errors() {
		codes[fieldErr.Field()] = fieldErr.Issue().Code()
	}

	assert.Equal(t, validation.CodeRequired, codes[fieldAddress])
	assert.Equal(t, validation.CodeRequired, codes[fieldDatabaseURL])
	assert.Equal(t, validation.CodeNotPositive, codes[fieldBcryptCost])
}

func TestLoadConfig_ReportsTheEnvironmentVariableKey(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv(envBcryptCost.Key, "0")
	t.Setenv(envAccessTokenTTL.Key, "-5m")

	_, err := LoadConfig()

	require.Error(t, err)
	assert.Contains(t, err.Error(), envBcryptCost.Key)
	assert.Contains(t, err.Error(), envAccessTokenTTL.Key)
	assert.Contains(t, err.Error(), validation.CodeNotPositive)
}

func TestLoadConfig_GroupsParseAndValidationErrors(t *testing.T) {
	t.Setenv(envBcryptCost.Key, "twelve")
	t.Setenv(envAccessTokenTTL.Key, "-5m")
	t.Setenv(envLogFormat.Key, "xml")
	t.Setenv(envRateLimit.Key, "extreme")

	_, err := LoadConfig()

	require.Error(t, err)
	msg := err.Error()
	assert.Contains(t, msg, envAddress.Key)
	assert.Contains(t, msg, envDatabaseURL.Key)
	assert.Contains(t, msg, envBcryptCost.Key)
	assert.Contains(t, msg, envAccessTokenTTL.Key)
	assert.Contains(t, msg, envLogFormat.Key)
	assert.Contains(t, msg, envRateLimit.Key)
	assert.Contains(t, msg, validation.CodeRequired)
	assert.Contains(t, msg, validation.CodeNotPositive)
	assert.Contains(t, msg, validation.CodeNotAllowed)
}
