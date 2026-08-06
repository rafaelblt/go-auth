package bootstrap

import (
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ReturnsConfig(t *testing.T) {
	expected := Config{
		Address:         "add ress vlaue",
		DatabaseURL:     "db url vlaue",
		BcryptCost:      8,
		AccessTokenTTL:  22 * time.Minute,
		RefreshTokenTTL: 2 * 24 * time.Hour,
	}

	t.Setenv(EnvAddress.Key, expected.Address)
	t.Setenv(EnvDatabaseURL.Key, expected.DatabaseURL)
	t.Setenv(EnvBcryptCost.Key, strconv.Itoa(expected.BcryptCost))
	t.Setenv(EnvAccessTokenTTL.Key, expected.AccessTokenTTL.String())
	t.Setenv(EnvRefreshTokenTTL.Key, expected.RefreshTokenTTL.String())

	actual, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestLoadConfig_ReturnsError_WhenNoEnvIsSet(t *testing.T) {
	cfg, err := LoadConfig()
	assert.Zero(t, cfg)
	assert.Error(t, err)
}
