package config

import (
	"net/netip"
	"time"
)

var envDatabaseURL = env[string]{
	Key:    "DATABASE_URL",
	Parser: stringEnvParser,
}

var envAddress = env[string]{
	Key:    "ADDRESS",
	Parser: stringEnvParser,
}

var envAutoMigrate = env[bool]{
	Key:    "AUTO_MIGRATE",
	Parser: boolEnvParser,
}

var envBcryptCost = env[int]{
	Key:    "BCRYPT_COST",
	Parser: intEnvParser,
	Presets: map[string]int{
		"FAST":   10,
		"NORMAL": defaultBcryptCost,
		"STRONG": 14,
	},
}

var envAccessTokenTTL = env[time.Duration]{
	Key:    "ACCESS_TOKEN_TTL",
	Parser: durationEnvParser,
	Presets: map[string]time.Duration{
		"SHORT":  10 * time.Minute,
		"NORMAL": defaultAccessTokenTTL,
		"LONG":   1 * time.Hour,
	},
}

var envRefreshTokenTTL = env[time.Duration]{
	Key:    "REFRESH_TOKEN_TTL",
	Parser: durationEnvParser,
	Presets: map[string]time.Duration{
		"SHORT":  24 * time.Hour,
		"NORMAL": defaultRefreshTokenTTL,
		"LONG":   30 * 24 * time.Hour,
	},
}

var envLogFormat = env[LogFormat]{
	Key:    "LOG_FORMAT",
	Parser: logFormatEnvParser,
}

var envRateLimit = env[RateLimitLevel]{
	Key:    "RATE_LIMIT",
	Parser: rateLimitLevelEnvParser,
}

var envTrustedProxies = env[[]netip.Prefix]{
	Key:    "TRUSTED_PROXIES",
	Parser: trustedProxiesEnvParser,
}

var envSigningKeyEncryptionKey = env[[]byte]{
	Key:    "SIGNING_KEY_ENCRYPTION_KEY",
	Parser: encryptionKeyEnvParser,
}
