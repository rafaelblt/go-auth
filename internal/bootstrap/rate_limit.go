package bootstrap

import (
	"log/slog"
	"time"

	"github.com/rafaelblt/go-auth/internal/api"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/port"
)

type rateLimits struct {
	register, login, refresh, changePassword port.RateLimit
}

var rateLimitsByLevel = map[config.RateLimitLevel]rateLimits{
	config.RateLimitRelaxed: {
		register:       port.RateLimit{Requests: 30, Period: time.Hour},
		login:          port.RateLimit{Requests: 30, Period: time.Minute},
		refresh:        port.RateLimit{Requests: 300, Period: time.Minute},
		changePassword: port.RateLimit{Requests: 30, Period: time.Hour},
	},
	config.RateLimitNormal: {
		register:       port.RateLimit{Requests: 10, Period: time.Hour},
		login:          port.RateLimit{Requests: 10, Period: time.Minute},
		refresh:        port.RateLimit{Requests: 100, Period: time.Minute},
		changePassword: port.RateLimit{Requests: 10, Period: time.Hour},
	},
	config.RateLimitStrict: {
		register:       port.RateLimit{Requests: 3, Period: time.Hour},
		login:          port.RateLimit{Requests: 3, Period: time.Minute},
		refresh:        port.RateLimit{Requests: 30, Period: time.Minute},
		changePassword: port.RateLimit{Requests: 3, Period: time.Hour},
	},
}

// newRateLimiting returns nil when rate limiting is off. A level missing from
// rateLimitsByLevel gets zero limits, which api.NewRouter rejects at startup.
func newRateLimiting(cfg config.Config, limiter port.RateLimiter) *api.RateLimiting {
	level := cfg.RateLimit()
	if level == config.RateLimitOff {
		return nil
	}

	limits := rateLimitsByLevel[level]
	rateLimiting := api.RateLimiting{
		Limiter:        limiter,
		TrustedProxies: cfg.TrustedProxies(),
		Register:       limits.register,
		Login:          limits.login,
		Refresh:        limits.refresh,
		ChangePassword: limits.changePassword,
	}
	return &rateLimiting
}

func logRateLimit(logger *slog.Logger, cfg config.Config) {
	trustedProxies := make([]string, 0)
	for _, prefix := range cfg.TrustedProxies() {
		trustedProxies = append(trustedProxies, prefix.String())
	}

	level := cfg.RateLimit()
	if level == config.RateLimitOff {
		logger.Warn("rate limiting is off; from v2 it cannot be turned off, only set to a level")
		if len(trustedProxies) > 0 {
			logger.Warn("trusted proxies ignored while rate limiting is off", "trusted_proxies", trustedProxies)
		}
		return
	}

	logger.Info("rate limiting on", "rate_limit", string(level), "trusted_proxies", trustedProxies)
}
