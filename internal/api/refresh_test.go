package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRefreshDecoder_ReturnsInput(t *testing.T) {
	body := refreshRequestBody{
		RefreshToken: "refresh token",
	}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	req := httptest.NewRequest("POST", "localhost:0510", bytes.NewReader(buf))

	in, err := refreshDecoder(req)

	require.NoError(t, err)
	assert.Equal(t, body.RefreshToken, in.RefreshToken)
}

func TestRefreshDecoder_ReturnsError_WhenRequestBodyIsNil(t *testing.T) {
	req := httptest.NewRequest("POST", "localhost:0510", nil)

	in, err := refreshDecoder(req)

	assert.Error(t, err)
	assert.Zero(t, in)
}

func TestRefreshEncoder_ReturnsResponse(t *testing.T) {
	output := refresh.Output{
		AccessToken: usecase.AccessTokenDTO{
			Value:     "raw access token value",
			ExpiresAt: time.Now().UTC().AddDate(500, 50, 5),
		},
		RefreshToken: usecase.RefreshTokenDTO{
			Value:     "raw refresh token value",
			ExpiresAt: time.Now().UTC().AddDate(800, 80, 8),
		},
	}

	resp := refreshEncoder(output)

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	require.IsType(t, refreshResponseBody{}, resp.Body)
	actualBody := resp.Body.(refreshResponseBody)
	expectedBody := refreshResponseBody{
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
