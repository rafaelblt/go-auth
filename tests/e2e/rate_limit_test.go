package e2e

import (
	"net/http"
	"strconv"
	"testing"

	"github.com/rafaelblt/go-auth/internal/config"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const TooManyRequestsCode = "too_many_requests"

// startRateLimitedApp starts a second app with rate limiting at level. The
// shared app stays at the default, off, since its limits would interfere with
// every other test.
func startRateLimitedApp(t *testing.T, level config.RateLimitLevel) *testutil.HTTPClient {
	t.Helper()
	return startSecondApp(t, config.ConfigParams{RateLimit: shared.Ptr(level)})
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
