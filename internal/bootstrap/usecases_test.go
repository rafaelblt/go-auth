package bootstrap

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func sessionRetentionConfigForTest(t *testing.T, retention config.SessionRetention) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(config.ConfigParams{
		Address:          "localhost:8080",
		DatabaseURL:      "postgres://localhost/test",
		SessionRetention: shared.Ptr(retention),
	})
	require.NoError(t, err)

	return cfg
}

// The session repository is never called here, so it needs no database.
func sessionCleanupInfraForTest() infraDeps {
	return infraDeps{
		Clock:    infra.NewSystemClock(),
		Sessions: &postgres.SessionRepo{},
	}
}

// TESTS

func TestBuildSessionCleanup_ReturnsNil_WhenRetentionIsOff(t *testing.T) {
	cfg := sessionRetentionConfigForTest(t, config.SessionRetentionOff())

	cleanup, err := buildSessionCleanup(cfg, sessionCleanupInfraForTest())

	require.NoError(t, err)
	assert.Nil(t, cleanup)
}

func TestBuildSessionCleanup_ReturnsTheUseCase_WhenRetentionIsSet(t *testing.T) {
	cfg := sessionRetentionConfigForTest(t, config.SessionRetentionOf(48*time.Hour))

	cleanup, err := buildSessionCleanup(cfg, sessionCleanupInfraForTest())

	require.NoError(t, err)
	assert.NotNil(t, cleanup)
}
