package bootstrap

import (
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ReturnsConfig(t *testing.T) {
	dbURL := "dburl"
	os.Setenv(EnvDatabaseURL, dbURL)

	cfg, err := LoadConfig()

	require.NoError(t, err)
	require.NotZero(t, cfg)
	assert.Equal(t, dbURL, cfg.DatabaseURL)
	assert.Equal(t, 8, cfg.BcryptCost)
	assert.Equal(t, time.Hour * 24 * 7, cfg.RefreshTokenTTL)
}

func TestLoadConfig_ReturnsError_WithDatabaseNotDefined(t *testing.T) {
	os.Unsetenv(EnvDatabaseURL)

	cfg, err := LoadConfig()

	assert.Zero(t, cfg)
	assert.Error(t, err)
}
