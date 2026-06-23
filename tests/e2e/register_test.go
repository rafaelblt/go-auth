package e2e

import (
	"net/http"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
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

const RegisterPath = "/v1/auth/register"

func TestRegister_SuccessResponse(t *testing.T) {
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

func TestRegister_UsernameAlreadyExistsResponse(t *testing.T) {
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
	assert.Equal(t, "USERNAME_ALREADY_EXISTS", respBody.Error.Code)
	assert.NotZero(t, respBody.Error.Message)
}

func TestRegister_ValidationErrorResponse(t *testing.T) {
	env := testApp.NewEnv(t)

	reqBody := RegisterRequestBody{
		Username: "r",
		Password: "1",
	}

	resp := env.Client.Post(t, RegisterPath, reqBody)

	require.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)

	respBody := DecodeBody[ValidationErrorResponseBody](t, resp)
	assert.Len(t, respBody.Errors, 2)

	firstErr := respBody.Errors[0]
	assert.Equal(t, "username", firstErr.Field)
	assert.Equal(t, "TOO_SHORT", firstErr.Code)
	assert.Equal(t, map[string]any{"min": float64(user.UsernameMinLen)}, firstErr.Details)

	secondErr := respBody.Errors[1]
	assert.Equal(t, "password", secondErr.Field)
	assert.Equal(t, "TOO_SHORT", secondErr.Code)
	assert.Equal(t, map[string]any{"min": float64(credential.PlainPasswordMinLen)}, secondErr.Details)
}
