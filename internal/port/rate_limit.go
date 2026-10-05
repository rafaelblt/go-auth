package port

import (
	"context"
	"time"
)

type RateLimiter interface {
	// Allow counts one request against key under limit and says whether it
	// was within it. A refused request is not counted.
	Allow(ctx context.Context, key string, limit RateLimit) (RateLimitDecision, error)
}

type RateLimit struct {
	Requests int
	Period   time.Duration
}

type RateLimitDecision struct {
	Allowed    bool
	RetryAfter time.Duration
}
