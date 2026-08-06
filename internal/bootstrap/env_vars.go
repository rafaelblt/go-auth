package bootstrap

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/bootstrap/env"
)

var EnvDatabaseURL = env.EnvVar[string]{
	Key:      "DATABASE_URL",
	Required: true,
	Parser:   env.StringParser,
}

var EnvAddress = env.EnvVar[string]{
	Key:      "ADDRESS",
	Required: true,
	Parser:   env.StringParser,
}

var EnvBcryptCost = env.EnvVar[int]{
	Key:     "BCRYPT_COST",
	Default: 12,
	Parser:  env.IntParser,
	Presets: map[string]int{
		"FAST":   10,
		"NORMAL": 12,
		"STRONG": 14,
	},
}

var EnvAccessTokenTTL = env.EnvVar[time.Duration]{
	Key:     "ACCESS_TOKEN_TTL",
	Default: 30 * time.Minute,
	Parser:  env.DurationParser,
	Presets: map[string]time.Duration{
		"SHORT":  10 * time.Minute,
		"NORMAL": 30 * time.Minute,
		"LONG":   1 * time.Hour,
	},
}

var EnvRefreshTokenTTL = env.EnvVar[time.Duration]{
	Key:     "REFRESH_TOKEN_TTL",
	Default: 7 * 24 * time.Hour,
	Parser:  env.DurationParser,
	Presets: map[string]time.Duration{
		"SHORT":  24 * time.Hour,
		"NORMAL": 7 * 24 * time.Hour,
		"LONG":   30 * 24 * time.Hour,
	},
}
