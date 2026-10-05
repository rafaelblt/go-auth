package ratelimit

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

var threeEvery30s = port.RateLimit{Requests: 3, Period: 30 * time.Second}

func inMemoryForTest(t *testing.T) (*InMemory, *porttest.FakeClock) {
	t.Helper()

	clock := porttest.NewFakeClock()
	limiter, err := NewInMemory(Config{Clock: clock})
	require.NoError(t, err)

	return limiter, clock
}

func allow(t *testing.T, limiter *InMemory, key string, limit port.RateLimit) port.RateLimitDecision {
	t.Helper()

	decision, err := limiter.Allow(t.Context(), key, limit)
	require.NoError(t, err)

	return decision
}

func useBurst(t *testing.T, limiter *InMemory, key string, limit port.RateLimit) {
	t.Helper()

	for range limit.Requests {
		require.True(t, allow(t, limiter, key, limit).Allowed)
	}
}

func advance(clock *porttest.FakeClock, d time.Duration) {
	clock.SetNow(clock.Now().Add(d))
}

// TESTS

func TestNewInMemory_ReturnsError_WhenClockIsNil(t *testing.T) {
	limiter, err := NewInMemory(Config{Clock: nil})

	require.Error(t, err)
	assert.Nil(t, limiter)
}

func TestInMemory_Allow_AllowsABurst_ThenRefusesWithRetryAfter(t *testing.T) {
	limiter, _ := inMemoryForTest(t)

	for range 3 {
		decision := allow(t, limiter, "a", threeEvery30s)
		assert.True(t, decision.Allowed)
		assert.Zero(t, decision.RetryAfter)
	}
	decision := allow(t, limiter, "a", threeEvery30s)

	assert.False(t, decision.Allowed)
	assert.Equal(t, 10*time.Second, decision.RetryAfter)
}

func TestInMemory_Allow_AllowsOneMore_AfterAnInterval(t *testing.T) {
	limiter, clock := inMemoryForTest(t)
	useBurst(t, limiter, "a", threeEvery30s)
	require.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)

	advance(clock, 10*time.Second)

	assert.True(t, allow(t, limiter, "a", threeEvery30s).Allowed)
	assert.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)
}

func TestInMemory_Allow_RestoresTheWholeBurst_AfterAPeriod(t *testing.T) {
	limiter, clock := inMemoryForTest(t)
	useBurst(t, limiter, "a", threeEvery30s)

	advance(clock, 30*time.Second)

	for range 3 {
		assert.True(t, allow(t, limiter, "a", threeEvery30s).Allowed)
	}
	assert.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)
}

func TestInMemory_Allow_DoesNotCountRefusedRequests(t *testing.T) {
	limiter, clock := inMemoryForTest(t)
	useBurst(t, limiter, "a", threeEvery30s)
	for range 5 {
		require.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)
	}

	advance(clock, 10*time.Second)

	assert.True(t, allow(t, limiter, "a", threeEvery30s).Allowed)
	assert.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)
}

func TestInMemory_Allow_KeepsKeysApart(t *testing.T) {
	limiter, _ := inMemoryForTest(t)
	useBurst(t, limiter, "a", threeEvery30s)
	require.False(t, allow(t, limiter, "a", threeEvery30s).Allowed)

	decision := allow(t, limiter, "b", threeEvery30s)

	assert.True(t, decision.Allowed)
}

func TestInMemory_Allow_ReturnsError_WhenLimitIsInvalid(t *testing.T) {
	testCases := []struct {
		desc  string
		limit port.RateLimit
	}{
		{
			desc:  "zero requests",
			limit: port.RateLimit{Requests: 0, Period: time.Minute},
		},
		{
			desc:  "negative requests",
			limit: port.RateLimit{Requests: -1, Period: time.Minute},
		},
		{
			desc:  "zero period",
			limit: port.RateLimit{Requests: 3, Period: 0},
		},
		{
			desc:  "negative period",
			limit: port.RateLimit{Requests: 3, Period: -time.Minute},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			limiter, _ := inMemoryForTest(t)

			decision, err := limiter.Allow(t.Context(), "a", tC.limit)

			require.ErrorIs(t, err, ErrInvalidLimit)
			assert.False(t, decision.Allowed)
		})
	}
}

func TestInMemory_Allow_ForgetsKeysWhoseAllowanceIsBack(t *testing.T) {
	limiter, clock := inMemoryForTest(t)
	allow(t, limiter, "a", threeEvery30s)
	allow(t, limiter, "b", threeEvery30s)

	advance(clock, threeEvery30s.Period+sweepInterval)
	allow(t, limiter, "c", threeEvery30s)

	assert.Len(t, limiter.tats, 1)
	assert.Contains(t, limiter.tats, "c")
}

func TestInMemory_Allow_IsSafeForConcurrentUse(t *testing.T) {
	limiter, _ := inMemoryForTest(t)
	limit := port.RateLimit{Requests: 100, Period: time.Minute}
	var allowed atomic.Int64

	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 10 {
				decision, err := limiter.Allow(t.Context(), "a", limit)
				if err == nil && decision.Allowed {
					allowed.Add(1)
				}
			}
		})
	}
	wg.Wait()

	assert.Equal(t, int64(100), allowed.Load())
}
