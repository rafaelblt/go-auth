package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type RegisterRequestBody struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponseBody struct {
	User User `json:"user"`
}

const (
	RegisterPath              = "/v1/auth/register"
	UsernameAlreadyExistsCode = "USERNAME_ALREADY_EXISTS"
	ValidationFailedCode      = "VALIDATION_FAILED"
)

func TestRegister_ReturnsSuccessResponse(t *testing.T) {
	env := testApp.NewEnv(t)

	reqBody := RegisterRequestBody{
		Username: "rafin",
		Password: "12345678",
	}

	resp := env.Client.Post(t, RegisterPath, reqBody)

	require.Equal(t, http.StatusOK, resp.StatusCode)
	respBody := DecodeBody[RegisterResponseBody](t, resp)
	assert.NotZero(t, respBody.User.ID)
	assert.Equal(t, reqBody.Username, respBody.User.Username)
	assert.Equal(t, "active", respBody.User.Status)
	assert.WithinRange(t, respBody.User.CreatedAt, time.Now().Add(-time.Second), time.Now().Add(time.Second))
	assert.Equal(t, respBody.User.CreatedAt, respBody.User.UpdatedAt)
}

func TestRegister_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, RegisterPath)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
}

func TestRegister_ReturnsUsernameAlreadyExistsErrorResponse(t *testing.T) {
	env := testApp.NewEnv(t)

	usr := usertest.NewUser(t, nil)
	env.Fixtures.SaveUser(t, usr)

	reqBody := RegisterRequestBody{
		Username: usr.Username().String(),
		Password: "12345678",
	}

	resp := env.Client.Post(t, RegisterPath, reqBody)

	require.Equal(t, http.StatusConflict, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, UsernameAlreadyExistsCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRegister_ReturnsValidationErrorResponse(t *testing.T) {
	env := testApp.NewEnv(t)
	reqBody := RegisterRequestBody{
		Username: "r",
		Password: "1",
	}

	resp := env.Client.Post(t, RegisterPath, reqBody)

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, ValidationFailedCode, respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
	require.Len(t, respBody.Error.Fields, 2)
	assert.Equal(t, "username", respBody.Error.Fields[0].Field)
	assert.Equal(t, "password", respBody.Error.Fields[1].Field)
}
