package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const UnsupportedMediaTypeCode = "UNSUPPORTED_MEDIA_TYPE"

func TestPostEndpoints_ReturnUnsupportedMediaType_WhenContentTypeIsNotJSON(t *testing.T) {
	endpoints := []struct {
		path string
		body any
	}{
		{path: RegisterPath, body: RegisterRequestBody{Username: "alice", Password: "correct-horse"}},
		{path: LoginPath, body: LoginRequestBody{Username: "alice", Password: "correct-horse"}},
		{path: RefreshPath, body: RefreshRequestBody{RefreshToken: "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01"}},
	}
	contentTypes := []string{"", "text/plain;charset=UTF-8", "application/x-www-form-urlencoded"}
	for _, endpoint := range endpoints {
		for _, contentType := range contentTypes {
			t.Run(endpoint.path+" "+contentType, func(t *testing.T) {
				env := testApp.NewEnv(t)

				resp := env.Client.PostWithContentType(t, endpoint.path, contentType, endpoint.body)

				require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)
				respBody := DecodeBody[ErrorResponseBody](t, resp)
				assert.Equal(t, UnsupportedMediaTypeCode, respBody.Error.Code)
			})
		}
	}
}

func TestRegister_CreatesNoUser_WhenContentTypeIsNotJSON(t *testing.T) {
	env := testApp.NewEnv(t)
	reqBody := RegisterRequestBody{
		Username: "alice",
		Password: "correct-horse",
	}

	resp := env.Client.PostWithContentType(t, RegisterPath, "text/plain;charset=UTF-8", reqBody)
	require.Equal(t, http.StatusUnsupportedMediaType, resp.StatusCode)

	resp = env.Client.Post(t, RegisterPath, reqBody)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestRegister_AcceptsJSONContentTypeWithParameters(t *testing.T) {
	env := testApp.NewEnv(t)
	reqBody := RegisterRequestBody{
		Username: "alice",
		Password: "correct-horse",
	}

	resp := env.Client.PostWithContentType(t, RegisterPath, "Application/JSON; charset=utf-8", reqBody)

	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestUnknownPath_ReturnsNotFound_WhenContentTypeIsNotJSON(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.PostWithContentType(t, "/v1/auth/unknown", "text/plain", "x")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
