package config

import (
	"time"
)

const (
	defaultAutoMigrate     = false
	defaultBcryptCost      = 12
	defaultAccessTokenTTL  = 30 * time.Minute
	defaultRefreshTokenTTL = 7 * 24 * time.Hour
	defaultLogFormat       = LogFormatJSON
)

var envDatabaseURL = env[string]{
	Key:      "DATABASE_URL",
	Required: true,
	Parser:   stringEnvParser,
}

var envAddress = env[string]{
	Key:      "ADDRESS",
	Required: true,
	Parser:   stringEnvParser,
}

var envAutoMigrate = env[bool]{
	Key:     "AUTO_MIGRATE",
	Default: defaultAutoMigrate,
	Parser:  boolEnvParser,
}

var envBcryptCost = env[int]{
	Key:     "BCRYPT_COST",
	Default: defaultBcryptCost,
	Parser:  intEnvParser,
	Presets: map[string]int{
		"FAST":   10,
		"NORMAL": defaultBcryptCost,
		"STRONG": 14,
	},
}

var envAccessTokenTTL = env[time.Duration]{
	Key:     "ACCESS_TOKEN_TTL",
	Default: defaultAccessTokenTTL,
	Parser:  durationEnvParser,
	Presets: map[string]time.Duration{
		"SHORT":  10 * time.Minute,
		"NORMAL": defaultAccessTokenTTL,
		"LONG":   1 * time.Hour,
	},
}

var envRefreshTokenTTL = env[time.Duration]{
	Key:     "REFRESH_TOKEN_TTL",
	Default: defaultRefreshTokenTTL,
	Parser:  durationEnvParser,
	Presets: map[string]time.Duration{
		"SHORT":  24 * time.Hour,
		"NORMAL": defaultRefreshTokenTTL,
		"LONG":   30 * 24 * time.Hour,
	},
}

var envLogFormat = env[LogFormat]{
	Key:     "LOG_FORMAT",
	Default: defaultLogFormat,
	Parser:  logFormatEnvParser,
}
