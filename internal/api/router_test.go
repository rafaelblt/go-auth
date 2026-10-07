package api

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"strings"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubUseCase[In, Out any] struct{}

func (stubUseCase[In, Out]) Execute(context.Context, In) (Out, error) {
	var out Out
	return out, nil
}

type panickingUseCase[In, Out any] struct{}

func (panickingUseCase[In, Out]) Execute(context.Context, In) (Out, error) {
	panic("boom")
}

func newRouterTestDependencies(t *testing.T) (Dependencies, *bytes.Buffer) {
	t.Helper()
	buf := &bytes.Buffer{}
	return Dependencies{
		Logger:            slog.New(slog.NewJSONHandler(buf, nil)),
		Clock:             porttest.NewFakeClock(),
		Register:          stubUseCase[register.Input, register.Output]{},
		Login:             stubUseCase[login.Input, login.Output]{},
		Refresh:           stubUseCase[refresh.Input, refresh.Output]{},
		Verify:            stubUseCase[verify.Input, verify.Output]{},
		PublicKeyProvider: NewFakePublicKeyProvider(),
	}, buf
}

func TestNewRouter_ReturnsError_WhenADependencyIsNil(t *testing.T) {
	testCases := []struct {
		desc     string
		nilOne   func(*Dependencies)
		expected string
	}{
		{desc: "logger", nilOne: func(d *Dependencies) { d.Logger = nil }, expected: "logger nil"},
		{desc: "clock", nilOne: func(d *Dependencies) { d.Clock = nil }, expected: "clock nil"},
		{desc: "register", nilOne: func(d *Dependencies) { d.Register = nil }, expected: "register nil"},
		{desc: "login", nilOne: func(d *Dependencies) { d.Login = nil }, expected: "login nil"},
		{desc: "refresh", nilOne: func(d *Dependencies) { d.Refresh = nil }, expected: "refresh nil"},
		{desc: "verify", nilOne: func(d *Dependencies) { d.Verify = nil }, expected: "verify nil"},
		{desc: "public key provider", nilOne: func(d *Dependencies) { d.PublicKeyProvider = nil }, expected: "public key provider nil"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			deps, _ := newRouterTestDependencies(t)
			tC.nilOne(&deps)

			handler, err := NewRouter(Config{Dependencies: deps})

			assert.EqualError(t, err, tC.expected)
			assert.Nil(t, handler)
		})
	}
}

func newRouterTestRateLimiting(limiter port.RateLimiter) *RateLimiting {
	return &RateLimiting{
		Limiter:  limiter,
		Register: port.RateLimit{Requests: 1, Period: time.Minute},
		Login:    port.RateLimit{Requests: 2, Period: time.Minute},
		Refresh:  port.RateLimit{Requests: 3, Period: time.Minute},
	}
}

func newRefusingRateLimiter(retryAfter time.Duration) *porttest.FakeRateLimiter {
	limiter := porttest.NewFakeRateLimiter()
	limiter.SetDecision(port.RateLimitDecision{Allowed: false, RetryAfter: retryAfter})
	return limiter
}

func newJSONPost(path string) *http.Request {
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	return req
}

func TestNewRouter_ReturnsError_WhenRateLimitingIsInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		mutate   func(*RateLimiting)
		expected string
	}{
		{desc: "limiter nil", mutate: func(rl *RateLimiting) { rl.Limiter = nil }, expected: "rate limiter nil"},
		{desc: "register requests zero", mutate: func(rl *RateLimiting) { rl.Register.Requests = 0 }, expected: "register rate limit invalid"},
		{desc: "login period zero", mutate: func(rl *RateLimiting) { rl.Login.Period = 0 }, expected: "login rate limit invalid"},
		{desc: "refresh requests negative", mutate: func(rl *RateLimiting) { rl.Refresh.Requests = -1 }, expected: "refresh rate limit invalid"},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			deps, _ := newRouterTestDependencies(t)
			rateLimiting := newRouterTestRateLimiting(porttest.NewFakeRateLimiter())
			tC.mutate(rateLimiting)

			handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: rateLimiting})

			assert.EqualError(t, err, tC.expected)
			assert.Nil(t, handler)
		})
	}
}

func TestNewRouter_RateLimitsEachPostEndpoint_WithItsOwnLimit(t *testing.T) {
	deps, _ := newRouterTestDependencies(t)
	limiter := newRefusingRateLimiter(time.Second)
	rateLimiting := newRouterTestRateLimiting(limiter)
	handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: rateLimiting})
	require.NoError(t, err)

	for _, path := range []string{"/v1/auth/register", "/v1/auth/login", "/v1/auth/refresh"} {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, newJSONPost(path))
		assert.Equal(t, http.StatusTooManyRequests, recorder.Code, path)
	}

	assert.Equal(t, []porttest.FakeRateLimiterCall{
		{Key: "register 192.0.2.1", Limit: rateLimiting.Register},
		{Key: "login 192.0.2.1", Limit: rateLimiting.Login},
		{Key: "refresh 192.0.2.1", Limit: rateLimiting.Refresh},
	}, limiter.Calls())
}

func TestNewRouter_CountsTheForwardedClient_WhenThePeerIsATrustedProxy(t *testing.T) {
	deps, _ := newRouterTestDependencies(t)
	limiter := porttest.NewFakeRateLimiter()
	rateLimiting := newRouterTestRateLimiting(limiter)
	rateLimiting.TrustedProxies = []netip.Prefix{netip.MustParsePrefix("10.0.0.0/8")}
	handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: rateLimiting})
	require.NoError(t, err)
	req := newJSONPost("/v1/auth/login")
	req.RemoteAddr = "10.0.0.1:1234"
	req.Header.Set("X-Forwarded-For", "203.0.113.9")

	handler.ServeHTTP(httptest.NewRecorder(), req)

	assert.Equal(t, []porttest.FakeRateLimiterCall{
		{Key: "login 203.0.113.9", Limit: rateLimiting.Login},
	}, limiter.Calls())
}

func TestNewRouter_DoesNotRateLimit_JWKSOrUnmatchedRoutes(t *testing.T) {
	deps, _ := newRouterTestDependencies(t)
	limiter := newRefusingRateLimiter(time.Second)
	handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: newRouterTestRateLimiting(limiter)})
	require.NoError(t, err)
	testCases := []struct {
		path     string
		expected int
	}{
		{path: "/.well-known/jwks.json", expected: http.StatusOK},
		{path: "/unknown", expected: http.StatusNotFound},
		{path: "/v1/auth/login", expected: http.StatusMethodNotAllowed},
	}

	for _, tC := range testCases {
		recorder := httptest.NewRecorder()
		handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, tC.path, nil))
		assert.Equal(t, tC.expected, recorder.Code, tC.path)
	}

	assert.Empty(t, limiter.Calls())
}

func TestNewRouter_DoesNotRateLimit_Verify(t *testing.T) {
	deps, _ := newRouterTestDependencies(t)
	limiter := newRefusingRateLimiter(time.Second)
	handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: newRouterTestRateLimiting(limiter)})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, newJSONPost("/v1/auth/verify"))

	assert.Equal(t, http.StatusOK, recorder.Code)
	assert.Empty(t, limiter.Calls())
}

func TestNewRouter_AnswersTooManyRequests_AndLogsItsLifecycle_WhenLimitIsExceeded(t *testing.T) {
	deps, buf := newRouterTestDependencies(t)
	limiter := newRefusingRateLimiter(1500 * time.Millisecond)
	handler, err := NewRouter(Config{Dependencies: deps, RateLimiting: newRouterTestRateLimiting(limiter)})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, newJSONPost("/v1/auth/login"))

	require.Equal(t, http.StatusTooManyRequests, recorder.Code)
	assert.Equal(t, "2", recorder.Header().Get("Retry-After"))
	lines := loggedLines(t, buf)
	require.Len(t, lines, 3)
	assert.Equal(t, "request received", lines[0]["msg"])
	assert.Equal(t, "rate limit exceeded", lines[1]["msg"])
	assert.Equal(t, lines[0]["request_id"], lines[1]["request_id"])
	assert.Equal(t, "request finished", lines[2]["msg"])
	assert.Equal(t, lines[0]["request_id"], lines[2]["request_id"])
	assert.Equal(t, float64(http.StatusTooManyRequests), lines[2]["status"])
}

func TestNewRouter_AnswersUnknownRoute_AndLogsItsLifecycle(t *testing.T) {
	deps, buf := newRouterTestDependencies(t)
	handler, err := NewRouter(Config{Dependencies: deps})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()

	handler.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/unknown", nil))

	require.Equal(t, http.StatusNotFound, recorder.Code)
	assert.Equal(t, "application/json", recorder.Header().Get("Content-Type"))
	assert.Equal(t, routeNotFoundError().Body, decodeErrorBody(t, recorder))
	lines := loggedLines(t, buf)
	require.NotEmpty(t, lines)
	assert.Equal(t, "request received", lines[0]["msg"])
	last := lines[len(lines)-1]
	assert.Equal(t, "request finished", last["msg"])
	assert.Equal(t, float64(http.StatusNotFound), last["status"])
	assert.Equal(t, float64(0), last["duration"])
}

func TestNewRouter_AnswersInternalServerError_AndLogsItsLifecycle_WhenHandlerPanics(t *testing.T) {
	deps, buf := newRouterTestDependencies(t)
	deps.Login = panickingUseCase[login.Input, login.Output]{}
	handler, err := NewRouter(Config{Dependencies: deps})
	require.NoError(t, err)
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/v1/auth/login", strings.NewReader(`{"username":"alice","password":"secret"}`))
	req.Header.Set("Content-Type", "application/json")

	handler.ServeHTTP(recorder, req)

	require.Equal(t, http.StatusInternalServerError, recorder.Code)
	assert.Equal(t, internalServerError().Body, decodeErrorBody(t, recorder))
	lines := loggedLines(t, buf)
	require.Len(t, lines, 3)
	assert.Equal(t, "request received", lines[0]["msg"])
	assert.Equal(t, "panic recovered", lines[1]["msg"])
	assert.Equal(t, lines[0]["request_id"], lines[1]["request_id"])
	assert.Equal(t, "request finished", lines[2]["msg"])
	assert.Equal(t, float64(http.StatusInternalServerError), lines[2]["status"])
}
