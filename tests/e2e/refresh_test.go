package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RefreshRequestBody struct {
	RefreshToken string `json:"refresh_token"`
}

type RefreshResponseBody struct {
	AccessToken  AccessToken  `json:"access_token"`
	RefreshToken RefreshToken `json:"refresh_token"`
}

const (
	RefreshPath      = "/v1/auth/refresh"
	InvalidTokenCode = "invalid_token"
)

func TestRefresh_ReturnsSuccessResponse(t *testing.T) {
	env := testApp.NewEnv(t)
	_, raw := env.Fixtures.CreateRefreshToken(t)
	reqBody := RefreshRequestBody{
		RefreshToken: raw,
	}

	resp := env.Client.Post(t, RefreshPath, reqBody)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[RefreshResponseBody](t, resp)
	assert.NotZero(t, respBody.AccessToken)
	assert.NotZero(t, respBody.RefreshToken)
	assert.GreaterOrEqual(t, respBody.AccessToken.ExpiresIn, int64((AccessTokenTTL-time.Second)/time.Second))
	assert.LessOrEqual(t, respBody.AccessToken.ExpiresIn, int64(AccessTokenTTL/time.Second))
	assert.Equal(t, int64(RefreshTokenTTL/time.Second), respBody.RefreshToken.ExpiresIn)
}

func TestRefresh_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, RefreshPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "POST", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}

func TestRefresh_ReturnsInvalidTokenResponse_WhenTokenIsInvalid(t *testing.T) {
	env := testApp.NewEnv(t)
	reqBody := RefreshRequestBody{
		RefreshToken: "invalid",
	}

	resp := env.Client.Post(t, RefreshPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRefresh_ReturnsInvalidTokenResponse_WhenTokenIsExpired(t *testing.T) {
	env := testApp.NewEnv(t)
	_, raw := env.Fixtures.CreateRefreshTokenExpired(t)

	reqBody := RefreshRequestBody{RefreshToken: raw}

	resp := env.Client.Post(t, RefreshPath, reqBody)
	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRefresh_ReturnsInvalidTokenResponse_WhenTokenIsAlreadyUsed(t *testing.T) {
	env := testApp.NewEnv(t)
	_, raw := env.Fixtures.CreateRefreshTokenAlreadyUsed(t)
	reqBody := RefreshRequestBody{RefreshToken: raw}

	resp := env.Client.Post(t, RefreshPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRefresh_ReturnsInvalidTokenResponse_WhenSessionIsRevoked(t *testing.T) {
	env := testApp.NewEnv(t)
	_, raw := env.Fixtures.CreateRefreshTokenAlreadyUsed(t)
	reqBody := RefreshRequestBody{RefreshToken: raw}

	resp := env.Client.Post(t, RefreshPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidTokenCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRefresh_RevokesSession_WhenTokenIsAlreadyUsed(t *testing.T) {
	env := testApp.NewEnv(t)
	token, raw := env.Fixtures.CreateRefreshTokenAlreadyUsed(t)
	reqBody := RefreshRequestBody{RefreshToken: raw}

	resp := env.Client.Post(t, RefreshPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	env.Asserts.RequireSessionIsRevoked(t, token.SessionID())
}
