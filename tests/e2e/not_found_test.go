package e2e

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNotFound(t *testing.T) {
	env := testApp.NewEnv(t)

	resp := env.Client.Get(t, "unknown")

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}
