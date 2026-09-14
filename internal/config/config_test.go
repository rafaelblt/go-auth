package config

import (
	"strconv"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
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

	t.Setenv(envAddress.Key, address)
	t.Setenv(envDatabaseURL.Key, databaseURL)
	t.Setenv(envAutoMigrate.Key, strconv.FormatBool(autoMigrate))
	t.Setenv(envBcryptCost.Key, strconv.Itoa(bcryptCost))
	t.Setenv(envAccessTokenTTL.Key, accessTokenTTL.String())
	t.Setenv(envRefreshTokenTTL.Key, refreshTokenTTL.String())

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, address, cfg.Address())
	assert.Equal(t, databaseURL, cfg.DatabaseURL())
	assert.Equal(t, autoMigrate, cfg.AutoMigrate())
	assert.Equal(t, bcryptCost, cfg.BcryptCost())
	assert.Equal(t, accessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, refreshTokenTTL, cfg.RefreshTokenTTL())
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
		AutoMigrate:     true,
		BcryptCost:      shared.Ptr(8),
		AccessTokenTTL:  shared.Ptr(22 * time.Minute),
		RefreshTokenTTL: shared.Ptr(2 * 24 * time.Hour),
	}

	cfg, err := NewConfig(params)

	require.NoError(t, err)
	assert.False(t, cfg.IsZero())
	assert.Equal(t, params.Address, cfg.Address())
	assert.Equal(t, params.DatabaseURL, cfg.DatabaseURL())
	assert.Equal(t, params.AutoMigrate, cfg.AutoMigrate())
	assert.Equal(t, *params.BcryptCost, cfg.BcryptCost())
	assert.Equal(t, *params.AccessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, *params.RefreshTokenTTL, cfg.RefreshTokenTTL())
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
}

func TestNewConfig_ReturnsError(t *testing.T) {
	valid := func() ConfigParams {
		return ConfigParams{
			Address:         "add ress vlaue",
			DatabaseURL:     "db url vlaue",
			BcryptCost:      shared.Ptr(8),
			AccessTokenTTL:  shared.Ptr(22 * time.Minute),
			RefreshTokenTTL: shared.Ptr(2 * 24 * time.Hour),
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
