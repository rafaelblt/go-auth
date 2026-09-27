package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	RouteNotFoundCode    = "ROUTE_NOT_FOUND"
	MethodNotAllowedCode = "METHOD_NOT_ALLOWED"
)

func TestNotFound(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, "unknown")

	require.Equal(t, http.StatusNotFound, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	respBody := DecodeBody[ErrorResponseBody](t, resp)
	assert.Equal(t, RouteNotFoundCode, respBody.Error.Code)
}
