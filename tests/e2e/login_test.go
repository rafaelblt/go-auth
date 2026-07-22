package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type LoginRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type LoginResponseBody struct {
	AccessToken  AccessToken  `json:"access_token"`
	RefreshToken RefreshToken `json:"refresh_token"`
}

const (
	LoginPath              = "/v1/auth/login"
	InvalidCredentialsCode = "INVALID_CREDENTIALS"
)

func TestLogin_ReturnsSuccessResponse(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	reqBody := LoginRequestBody{
		Username: usr.Username().String(),
		Password: pwd.Value(),
	}

	resp := env.Client.Post(t, LoginPath, reqBody)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[LoginResponseBody](t, resp)
	assert.NotZero(t, respBody.AccessToken)
	assert.NotZero(t, respBody.RefreshToken)
}

func TestLogin_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, LoginPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestLogin_ReturnsInvalidCredentialsResponse_WhenUsernameNotExists(t *testing.T) {
	env := testApp.NewEnv(t)
	reqBody := LoginRequestBody{
		Username: "rafaelblt",
		Password: "12345678",
	}

	resp := env.Client.Post(t, LoginPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidCredentialsCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestLogin_ReturnsInvalidCredentialsResponse_WhenPasswordIsIncorrect(t *testing.T) {
	env := testApp.NewEnv(t)
	usr, pwd := env.Fixtures.CreateUserAndPassword(t)
	reqBody := LoginRequestBody{
		Username: usr.Username().String(),
		Password: pwd.Value() + "X", // incorrect
	}

	resp := env.Client.Post(t, LoginPath, reqBody)

	require.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, InvalidCredentialsCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}
