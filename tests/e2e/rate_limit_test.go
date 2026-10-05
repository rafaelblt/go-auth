package e2e

import (
	"context"
	"net/http"
	"strconv"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/bootstrap"
	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TooManyRequestsCode = "too_many_requests"

// startRateLimitedApp starts a second app, on its own address and with
// rate limiting at level, over the suite's database. The shared app stays at
// the default, off, since its limits would interfere with every other test.
func startRateLimitedApp(t *testing.T, level config.RateLimitLevel) *testutil.HTTPClient {
	t.Helper()

	cfg, err := config.NewConfig(config.ConfigParams{
		Address:     "localhost:8081",
		DatabaseURL: testApp.cfg.DatabaseURL(),
		BcryptCost:  shared.Ptr(6),
		RateLimit:   shared.Ptr(level),
	})
	require.NoError(t, err)

	ctx, cancel := context.WithCancel(t.Context())
	app, err := bootstrap.NewApp(ctx, bootstrap.AppParams{Config: cfg})
	require.NoError(t, err)

	var runErr error
	stopped := make(chan struct{})
	go func() {
		runErr = app.Run(ctx)
		close(stopped)
	}()
	t.Cleanup(func() {
		cancel()
		<-stopped
		app.Close()
		assert.NoError(t, runErr)
	})

	baseURL := "http://" + cfg.Address()
	deadline := time.After(5 * time.Second)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for {
		resp, err := http.Get(baseURL + JWKSPath)
		if err == nil {
			resp.Body.Close()
			break
		}
		select {
		case <-stopped:
			t.Fatalf("rate limited app stopped before serving: %v", runErr)
		case <-deadline:
			t.Fatalf("rate limited app not serving after 5s: %v", err)
		case <-ticker.C:
		}
	}

	client, err := testutil.NewHTTPClient(baseURL)
	require.NoError(t, err)
	return client
}

func TestLogin_ReturnsTooManyRequests_WhenStrictLimitIsUsedUp(t *testing.T) {
	testApp.NewEnv(t)
	client := startRateLimitedApp(t, config.RateLimitStrict)
	reqBody := LoginRequestBody{
		Username: "rafaelblt",
		Password: "12345678",
	}
	for range 3 {
		resp := client.Post(t, LoginPath, reqBody)
		require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}

	resp := client.Post(t, LoginPath, reqBody)

	require.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, TooManyRequestsCode, respBody.Error.Code)
	retryAfter, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	require.NoError(t, err)
	assert.GreaterOrEqual(t, retryAfter, 1)
	assert.LessOrEqual(t, retryAfter, 20)

	resp = client.PostWithContentType(t, LoginPath, "text/plain;charset=UTF-8", reqBody)
	require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
	assert.Equal(t, UnsupportedMediaTypeCode, DecodeBody[ErrorResponseBody](t, resp).Error.Code)

	resp = client.Get(t, JWKSPath)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp = client.Post(t, RegisterPath, RegisterRequestBody{
		Username: "alice",
		Password: "correct-horse",
	})
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
