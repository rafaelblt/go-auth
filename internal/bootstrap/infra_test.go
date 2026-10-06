package bootstrap

import (
	"bytes"
	"testing"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func signingKeyEncryptionConfigForTest(t *testing.T, encryptionKey []byte) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(config.ConfigParams{
		Address:                 "localhost:8080",
		DatabaseURL:             "postgres://localhost/test",
		SigningKeyEncryptionKey: encryptionKey,
	})
	require.NoError(t, err)

	return cfg
}

// TESTS

func TestLogSigningKeyEncryption_WarnsThatV2RequiresIt_WhenUnset(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	cfg := signingKeyEncryptionConfigForTest(t, nil)

	logSigningKeyEncryption(logger, cfg)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "WARN", line["level"])
	assert.Equal(t, "signing keys are stored unencrypted; from v2 SIGNING_KEY_ENCRYPTION_KEY is required", line["msg"])
}

func TestLogSigningKeyEncryption_LogsNothing_WhenSet(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	cfg := signingKeyEncryptionConfigForTest(t, bytes.Repeat([]byte{0x2a}, 32))

	logSigningKeyEncryption(logger, cfg)

	assert.Empty(t, loggedLines(t, buf))
}
