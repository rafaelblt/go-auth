package bootstrap

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/netip"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func rateLimitConfigForTest(t *testing.T, level config.RateLimitLevel, trustedProxies []netip.Prefix) config.Config {
	t.Helper()

	cfg, err := config.NewConfig(config.ConfigParams{
		Address:        "localhost:8080",
		DatabaseURL:    "postgres://localhost/test",
		RateLimit:      shared.Ptr(level),
		TrustedProxies: trustedProxies,
	})
	require.NoError(t, err)

	return cfg
}

func loggerWithLoggedLines() (*slog.Logger, *bytes.Buffer) {
	buf := &bytes.Buffer{}
	return slog.New(slog.NewJSONHandler(buf, nil)), buf
}

func loggedLines(t *testing.T, buf *bytes.Buffer) []map[string]any {
	t.Helper()

	lines := []map[string]any{}
	decoder := json.NewDecoder(buf)
	for decoder.More() {
		line := map[string]any{}
		require.NoError(t, decoder.Decode(&line))
		lines = append(lines, line)
	}
	return lines
}

// TESTS

func TestNewRateLimiting_ReturnsNil_WhenOff(t *testing.T) {
	cfg := rateLimitConfigForTest(t, config.RateLimitOff, nil)

	rateLimiting := newRateLimiting(cfg, porttest.NewFakeRateLimiter())

	assert.Nil(t, rateLimiting)
}

func TestNewRateLimiting_ReturnsTheLevelsLimits(t *testing.T) {
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	cfg := rateLimitConfigForTest(t, config.RateLimitNormal, trusted)
	limiter := porttest.NewFakeRateLimiter()

	rateLimiting := newRateLimiting(cfg, limiter)

	require.NotNil(t, rateLimiting)
	assert.Same(t, limiter, rateLimiting.Limiter)
	assert.Equal(t, trusted, rateLimiting.TrustedProxies)
	assert.Equal(t, port.RateLimit{Requests: 10, Period: time.Hour}, rateLimiting.Register)
	assert.Equal(t, port.RateLimit{Requests: 10, Period: time.Minute}, rateLimiting.Login)
	assert.Equal(t, port.RateLimit{Requests: 100, Period: time.Minute}, rateLimiting.Refresh)
}

// The numbers are documented per level in docs/configuration.md; a change here
// is a change there.
func TestRateLimitsByLevel_MatchesTheDocumentedTable(t *testing.T) {
	expected := map[config.RateLimitLevel]rateLimits{
		config.RateLimitRelaxed: {
			register: port.RateLimit{Requests: 30, Period: time.Hour},
			login:    port.RateLimit{Requests: 30, Period: time.Minute},
			refresh:  port.RateLimit{Requests: 300, Period: time.Minute},
		},
		config.RateLimitNormal: {
			register: port.RateLimit{Requests: 10, Period: time.Hour},
			login:    port.RateLimit{Requests: 10, Period: time.Minute},
			refresh:  port.RateLimit{Requests: 100, Period: time.Minute},
		},
		config.RateLimitStrict: {
			register: port.RateLimit{Requests: 3, Period: time.Hour},
			login:    port.RateLimit{Requests: 3, Period: time.Minute},
			refresh:  port.RateLimit{Requests: 30, Period: time.Minute},
		},
	}

	assert.Equal(t, expected, rateLimitsByLevel)
}

func TestLogRateLimit_WarnsThatV2CannotTurnItOff_WhenOff(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	cfg := rateLimitConfigForTest(t, config.RateLimitOff, nil)

	logRateLimit(logger, cfg)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "WARN", line["level"])
	assert.Equal(t, "rate limiting is off; from v2 it cannot be turned off, only set to a level", line["msg"])
}

func TestLogRateLimit_LogsLevelAndTrustedProxies_WhenOn(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	trusted := []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	cfg := rateLimitConfigForTest(t, config.RateLimitStrict, trusted)

	logRateLimit(logger, cfg)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "INFO", line["level"])
	assert.Equal(t, "rate limiting on", line["msg"])
	assert.Equal(t, "strict", line["rate_limit"])
	assert.Equal(t, []any{"10.0.0.0/8"}, line["trusted_proxies"])
}
