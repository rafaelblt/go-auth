package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/apitest"
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
	output := login.Output{
		AccessToken:  apitest.NewAccessTokenDTO(t),
		RefreshToken: apitest.NewRefreshTokenDTO(t),
	}

	resp := loginEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, loginResponseBody{}, resp.Body)
	actualBody := resp.Body.(loginResponseBody)
	expectedBody := loginResponseBody{
		AccessToken: accessToken{
			Value:     output.AccessToken.Value(),
			ExpiresAt: output.AccessToken.ExpiresAt(),
			ExpiresIn: int64(output.AccessToken.ExpiresIn() / time.Second),
		},
		RefreshToken: refreshToken{
			Value:     output.RefreshToken.Value(),
			ExpiresAt: output.RefreshToken.ExpiresAt(),
			ExpiresIn: int64(output.RefreshToken.ExpiresIn() / time.Second),
		},
	}
	assert.Equal(t, expectedBody, actualBody)
}
