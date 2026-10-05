package porttest

import (
	"context"

	"github.com/rafaelblt/go-auth/internal/port"
)

type FakeRateLimiter struct {
	calls    []FakeRateLimiterCall
	decision port.RateLimitDecision
	err      error
}

type FakeRateLimiterCall struct {
	Key   string
	Limit port.RateLimit
}

func NewFakeRateLimiter() *FakeRateLimiter {
	limiter := FakeRateLimiter{
		decision: port.RateLimitDecision{Allowed: true},
	}
	return &limiter
}

func (limiter *FakeRateLimiter) Allow(
	ctx context.Context, key string, limit port.RateLimit,
) (port.RateLimitDecision, error) {
	limiter.calls = append(limiter.calls, FakeRateLimiterCall{
		Key:   key,
		Limit: limit,
	})

	if limiter.err != nil {
		return port.RateLimitDecision{}, limiter.err
	}
	return limiter.decision, nil
}

func (limiter *FakeRateLimiter) SetDecision(decision port.RateLimitDecision) {
	limiter.decision = decision
}

func (limiter *FakeRateLimiter) SetError(err error) {
	limiter.err = err
}

func (limiter *FakeRateLimiter) Calls() []FakeRateLimiterCall {
	return limiter.calls
}
