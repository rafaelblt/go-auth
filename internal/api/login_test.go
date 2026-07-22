package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginDecoder_ReturnsInput(t *testing.T) {
	body := loginRequestBody{
		Username: "username",
		Password: "password",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:2040", bytes.NewReader(buf))

	in, err := loginDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.Username, in.Username)
	assert.Equal(t, body.Password, in.Password)
}

func TestLoginDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:2040", nil)

	in, err := loginDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestLoginEncoder_ReturnsResponse(t *testing.T) {
	recorder := httptest.NewRecorder()
	output := login.Output{
		AccessToken: usecase.AccessTokenDTO{
			Value:     "raw access token value",
			ExpiresAt: time.Now().UTC().AddDate(100, 10, 1),
		},
		RefreshToken: usecase.RefreshTokenDTO{
			Value:     "raw refresh token value",
			ExpiresAt: time.Now().UTC().AddDate(200, 20, 2),
		},
	}

	resp := loginEncoder(output)

	assert.Equal(t, http.StatusOK, recorder.Code)
	require.IsType(t, loginResponseBody{}, resp.Body)
	actualBody := resp.Body.(loginResponseBody)
	expectedBody := loginResponseBody{
		AccessToken: accessToken{
			Value:     output.AccessToken.Value,
			ExpiresAt: output.AccessToken.ExpiresAt,
		},
		RefreshToken: refreshToken{
			Value:     output.RefreshToken.Value,
			ExpiresAt: output.RefreshToken.ExpiresAt,
		},
	}
	assert.Equal(t, expectedBody, actualBody)
}
