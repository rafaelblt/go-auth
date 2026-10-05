// Package ratelimit counts requests per key for port.RateLimiter. InMemory
// keeps the counts in this process only.
//
// See docs/architecture/http.md#rate-limiting.
package ratelimit

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

var ErrInvalidLimit = errors.New("rate limit invalid")

const sweepInterval = time.Minute

type Config struct {
	Clock port.Clock
}

type InMemory struct {
	mu    sync.Mutex
	clock port.Clock
	// tats holds, per key, the time at which the key has its whole allowance
	// back.
	tats      map[string]time.Time
	nextSweep time.Time
}

func NewInMemory(cfg Config) (*InMemory, error) {
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}

	limiter := InMemory{
		clock: cfg.Clock,
		tats:  make(map[string]time.Time),
	}
	return &limiter, nil
}

func (m *InMemory) Allow(
	ctx context.Context, key string, limit port.RateLimit,
) (port.RateLimitDecision, error) {
	if limit.Requests <= 0 || limit.Period <= 0 {
		return port.RateLimitDecision{}, ErrInvalidLimit
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	now := m.clock.Now()
	m.sweep(now)

	interval := limit.Period / time.Duration(limit.Requests)
	tat, ok := m.tats[key]
	if !ok || tat.Before(now) {
		tat = now
	}

	ahead := tat.Sub(now)
	maxAhead := limit.Period - interval
	if ahead > maxAhead {
		return port.RateLimitDecision{RetryAfter: ahead - maxAhead}, nil
	}

	m.tats[key] = tat.Add(interval)
	return port.RateLimitDecision{Allowed: true}, nil
}

func (m *InMemory) sweep(now time.Time) {
	if now.Before(m.nextSweep) {
		return
	}
	for key, tat := range m.tats {
		if !tat.After(now) {
			delete(m.tats, key)
		}
	}
	m.nextSweep = now.Add(sweepInterval)
}
