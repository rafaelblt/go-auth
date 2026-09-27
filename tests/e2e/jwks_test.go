package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type JWK struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	X   string `json:"x"`
	Use string `json:"use"`
	Alg string `json:"alg"`
	Kid string `json:"kid"`
}

type JWKSResponseBody struct {
	Keys []JWK `json:"keys"`
}

const JWKSPath = "/.well-known/jwks.json"

func TestJWKS_ReturnsResponse(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, JWKSPath)

	body := DecodeBody[JWKSResponseBody](t, resp)
	require.NotNil(t, body.Keys)
	assert.Len(t, body.Keys, 1)
}

func TestJWKS_ReturnsMethodNotAllowed(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Post(t, JWKSPath, nil)

	require.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)
	assert.Equal(t, "GET, HEAD", resp.Header.Get("Allow"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, MethodNotAllowedCode, respBody.Error.Code)
}
