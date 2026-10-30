package config

import (
	"bytes"
	"encoding/base64"
	"net/netip"
	"strconv"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_ReturnsConfig(t *testing.T) {
	address := "localhost:8080"
	databaseURL := "postgres://user:password@localhost:5432/db"
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
	t.Setenv(envSigningKeyEncryptionKey.Key, base64.StdEncoding.EncodeToString(encryptionKeyForTest))

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
	assert.Equal(t, encryptionKeyForTest, cfg.SigningKeyEncryptionKey())
}

func TestLoadConfig_ReturnsDefaults_WhenOptionalEnvsAreMissing(t *testing.T) {
	t.Setenv(envAddress.Key, "localhost:8080")
	t.Setenv(envDatabaseURL.Key, "postgres://user:password@localhost:5432/db")

	cfg, err := LoadConfig()

	require.NoError(t, err)
	assert.Equal(t, defaultAutoMigrate, cfg.AutoMigrate())
	assert.Equal(t, defaultBcryptCost, cfg.BcryptCost())
	assert.Equal(t, defaultAccessTokenTTL, cfg.AccessTokenTTL())
	assert.Equal(t, defaultRefreshTokenTTL, cfg.RefreshTokenTTL())
	assert.Equal(t, defaultLogFormat, cfg.LogFormat())
	assert.Equal(t, RateLimitOff, cfg.RateLimit())
	assert.Empty(t, cfg.TrustedProxies())
	assert.Nil(t, cfg.SigningKeyEncryptionKey())
}

func setRequiredEnvs(t *testing.T) {
	t.Helper()
	t.Setenv(envAddress.Key, "localhost:8080")
	t.Setenv(envDatabaseURL.Key, "postgres://user:password@localhost:5432/db")
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
			desc: "signing key encryption key not 32 bytes",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envSigningKeyEncryptionKey.Key, base64.StdEncoding.EncodeToString(encryptionKeyForTest[:16]))
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
		{
			desc: "address without a port",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envAddress.Key, "8080")
			},
		},
		{
			desc: "database url blank",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envDatabaseURL.Key, "   ")
			},
		},
		{
			desc: "database url not a connection string",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envDatabaseURL.Key, "db.internal")
			},
		},
		{
			desc: "bcrypt cost above bcrypt's maximum",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envBcryptCost.Key, "32")
			},
		},
		{
			desc: "access token ttl over a day",
			setup: func(t *testing.T) {
				setRequiredEnvs(t)
				t.Setenv(envAccessTokenTTL.Key, "25h")
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
		Address:         "localhost:8080",
		DatabaseURL:     "postgres://user:password@localhost:5432/db",
		AutoMigrate:     shared.Ptr(true),
		BcryptCost:      shared.Ptr(8),
		AccessTokenTTL:  shared.Ptr(22 * time.Minute),
		RefreshTokenTTL: shared.Ptr(2 * 24 * time.Hour),
		LogFormat:       shared.Ptr(LogFormatText),
		RateLimit:       shared.Ptr(RateLimitNormal),
		TrustedProxies:  []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")},

		SigningKeyEncryptionKey: encryptionKeyForTest,
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
	assert.Equal(t, params.SigningKeyEncryptionKey, cfg.SigningKeyEncryptionKey())
}

func TestNewConfig_ReturnsDefaults_WhenOptionalParamsAreMissing(t *testing.T) {
	params := ConfigParams{
		Address:     "localhost:8080",
		DatabaseURL: "postgres://user:password@localhost:5432/db",
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
	assert.Nil(t, cfg.SigningKeyEncryptionKey())
}

func TestNewConfig_TreatsAnEmptySigningKeyEncryptionKeyAsNotSet(t *testing.T) {
	params := ConfigParams{
		Address:                 "localhost:8080",
		DatabaseURL:             "postgres://user:password@localhost:5432/db",
		SigningKeyEncryptionKey: []byte{},
	}

	cfg, err := NewConfig(params)

	require.NoError(t, err)
	assert.Nil(t, cfg.SigningKeyEncryptionKey())
}

func TestNewConfig_ReturnsError(t *testing.T) {
	valid := func() ConfigParams {
		return ConfigParams{
			Address:         "localhost:8080",
			DatabaseURL:     "postgres://user:password@localhost:5432/db",
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
		{
			desc:   "address without a port",
			mutate: func(p *ConfigParams) { p.Address = "localhost" },
		},
		{
			desc:   "address with a port out of range",
			mutate: func(p *ConfigParams) { p.Address = "localhost:99999" },
		},
		{
			desc:   "database url blank",
			mutate: func(p *ConfigParams) { p.DatabaseURL = " " },
		},
		{
			desc:   "database url with a bad port",
			mutate: func(p *ConfigParams) { p.DatabaseURL = "postgres://localhost:abc/db" },
		},
		{
			desc:   "bcrypt cost below bcrypt's minimum",
			mutate: func(p *ConfigParams) { p.BcryptCost = shared.Ptr(3) },
		},
		{
			desc:   "access token ttl under a minute",
			mutate: func(p *ConfigParams) { p.AccessTokenTTL = shared.Ptr(30 * time.Second) },
		},
		{
			desc:   "signing key encryption key of 16 bytes",
			mutate: func(p *ConfigParams) { p.SigningKeyEncryptionKey = encryptionKeyForTest[:16] },
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
		Address:        "localhost:8080",
		DatabaseURL:    "postgres://user:password@localhost:5432/db",
		TrustedProxies: []netip.Prefix{trusted},
	}
	cfg, err := NewConfig(params)
	require.NoError(t, err)

	params.TrustedProxies[0] = netip.MustParsePrefix("0.0.0.0/0")
	cfg.TrustedProxies()[0] = netip.MustParsePrefix("0.0.0.0/0")

	assert.Equal(t, []netip.Prefix{trusted}, cfg.TrustedProxies())
}

func TestConfig_SigningKeyEncryptionKey_IsACopy(t *testing.T) {
	params := ConfigParams{
		Address:                 "localhost:8080",
		DatabaseURL:             "postgres://user:password@localhost:5432/db",
		SigningKeyEncryptionKey: bytes.Clone(encryptionKeyForTest),
	}
	cfg, err := NewConfig(params)
	require.NoError(t, err)

	params.SigningKeyEncryptionKey[0] ^= 0xff
	cfg.SigningKeyEncryptionKey()[0] ^= 0xff

	assert.Equal(t, encryptionKeyForTest, cfg.SigningKeyEncryptionKey())
}

func TestNewConfig_NamesEachProblemByItsVariable(t *testing.T) {
	params := ConfigParams{BcryptCost: shared.Ptr(0)}

	_, err := NewConfig(params)

	assert.Equal(t, map[string]string{
		envAddress.Key:     "must be set",
		envDatabaseURL.Key: "must be set",
		envBcryptCost.Key:  "must be between 4 and 31, not 0",
	}, problemsOf(t, err))
}

func TestLoadConfig_ReportsTheEnvironmentVariableKey(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv(envBcryptCost.Key, "0")
	t.Setenv(envAccessTokenTTL.Key, "-5m")

	_, err := LoadConfig()

	assert.Equal(t, map[string]string{
		envBcryptCost.Key:     "must be between 4 and 31, not 0",
		envAccessTokenTTL.Key: "must be between 1m and 24h, not -5m",
	}, problemsOf(t, err))
}

func TestLoadConfig_ReportsTheEncryptionKeyVariable_WithoutItsValue(t *testing.T) {
	testCases := []struct {
		desc  string
		value string
	}{
		{
			desc:  "not base64",
			value: "not base64!",
		},
		{
			desc:  "base64 of 16 bytes",
			value: base64.StdEncoding.EncodeToString(encryptionKeyForTest[:16]),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			setRequiredEnvs(t)
			t.Setenv(envSigningKeyEncryptionKey.Key, tC.value)

			_, err := LoadConfig()

			require.Error(t, err)
			assert.Contains(t, err.Error(), envSigningKeyEncryptionKey.Key)
			assert.NotContains(t, err.Error(), tC.value)
		})
	}
}

func TestLoadConfig_ReportsParseAndValueProblemsTogether(t *testing.T) {
	t.Setenv(envAddress.Key, "")
	t.Setenv(envDatabaseURL.Key, "")
	t.Setenv(envBcryptCost.Key, "twelve")
	t.Setenv(envAccessTokenTTL.Key, "-5m")
	t.Setenv(envLogFormat.Key, "xml")
	t.Setenv(envRateLimit.Key, "extreme")

	_, err := LoadConfig()

	require.Error(t, err)
	assert.Equal(t, `invalid configuration, 6 settings to fix:
  ACCESS_TOKEN_TTL  must be between 1m and 24h, not -5m
  ADDRESS           must be set
  BCRYPT_COST       must be a whole number, not "twelve"
  DATABASE_URL      must be set
  LOG_FORMAT        must be json or text, not "xml"
  RATE_LIMIT        must be off, relaxed, normal or strict, not "extreme"`, err.Error())
}

// The checks that NewApp used to make one at a time, after connecting to the
// database, are made with the others.
func TestLoadConfig_ReportsWhatTheComponentsWouldRefuse_WithTheOtherSettings(t *testing.T) {
	t.Setenv(envAddress.Key, "8080")
	t.Setenv(envDatabaseURL.Key, "   ")
	t.Setenv(envBcryptCost.Key, "40")
	t.Setenv(envAccessTokenTTL.Key, "25h")
	t.Setenv(envLogFormat.Key, "xml")

	_, err := LoadConfig()

	require.Error(t, err)
	assert.Equal(t, `invalid configuration, 5 settings to fix:
  ACCESS_TOKEN_TTL  must be between 1m and 24h, not 25h
  ADDRESS           must be host:port, such as localhost:8080 (address 8080: missing port in address)
  BCRYPT_COST       must be between 4 and 31, not 40
  DATABASE_URL      must not be blank
  LOG_FORMAT        must be json or text, not "xml"`, err.Error())
}

func TestLoadConfig_ReportsTheDatabaseURLVariable_WithoutItsValue(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv(envDatabaseURL.Key, "postgres://alice:s3cret@db.internal:abc/app")

	_, err := LoadConfig()

	require.Error(t, err)
	assert.Equal(t, map[string]string{
		envDatabaseURL.Key: `must be a PostgreSQL connection string: failed to parse as URL (invalid port ":abc" after host)`,
	}, problemsOf(t, err))
	assert.NotContains(t, err.Error(), "s3cret")
	assert.NotContains(t, err.Error(), "db.internal")
}

func TestLoadConfig_SessionRetention(t *testing.T) {
	testCases := []struct {
		desc     string
		value    *string
		expected SessionRetention
	}{
		{desc: "not set", value: nil, expected: SessionRetentionOf(defaultSessionRetention)},
		{desc: "off", value: shared.Ptr("off"), expected: SessionRetentionOff()},
		{desc: "OFF", value: shared.Ptr("OFF"), expected: SessionRetentionOff()},
		{desc: "an hour", value: shared.Ptr("1h"), expected: SessionRetentionOf(time.Hour)},
		{desc: "two days", value: shared.Ptr("48h"), expected: SessionRetentionOf(48 * time.Hour)},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			setRequiredEnvs(t)
			if tC.value != nil {
				t.Setenv(envSessionRetention.Key, *tC.value)
			}

			cfg, err := LoadConfig()

			require.NoError(t, err)
			assert.Equal(t, tC.expected, cfg.SessionRetention())
		})
	}
}

func TestLoadConfig_RejectsASessionRetentionThatIsNotOffOrAtLeastAnHour(t *testing.T) {
	for _, value := range []string{"59m", "0", "0s", "-1h", "30d", "never"} {
		t.Run(value, func(t *testing.T) {
			setRequiredEnvs(t)
			t.Setenv(envSessionRetention.Key, value)

			cfg, err := LoadConfig()

			require.Error(t, err)
			assert.Contains(t, err.Error(), envSessionRetention.Key)
			assert.Zero(t, cfg)
		})
	}
}

func TestLoadConfig_ReportsTheSessionRetentionMinimum(t *testing.T) {
	setRequiredEnvs(t)
	t.Setenv(envSessionRetention.Key, "30m")

	_, err := LoadConfig()

	assert.Equal(t, map[string]string{
		envSessionRetention.Key: "must be at least 1h, not 30m",
	}, problemsOf(t, err))
}

func TestNewConfig_RejectsASessionRetentionUnderAnHour(t *testing.T) {
	for _, retention := range []SessionRetention{{}, SessionRetentionOf(0), SessionRetentionOf(59 * time.Minute)} {
		cfg, err := NewConfig(ConfigParams{
			Address:          "localhost:8080",
			DatabaseURL:      "postgres://user:password@localhost:5432/db",
			SessionRetention: &retention,
		})

		assert.Error(t, err)
		assert.True(t, cfg.IsZero())
	}
}

// problemsOf maps each variable err names to what it says about it.
func problemsOf(t *testing.T, err error) map[string]string {
	t.Helper()

	var cfgErr *configError
	require.ErrorAs(t, err, &cfgErr)

	problems := make(map[string]string)
	for _, p := range cfgErr.problems {
		problems[p.key] = p.err.Error()
	}
	return problems
}
