package e2e

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const RequestBodyTooLargeCode = "request_body_too_large"

const requestBodyMaxBytes = 64 << 10

func TestPostEndpoints_ReturnRequestBodyTooLarge_WhenBodyExceedsLimit(t *testing.T) {
	body := `{"username":"` + strings.Repeat("a", requestBodyMaxBytes) + `"}`
	for _, path := range []string{RegisterPath, LoginPath, RefreshPath} {
		t.Run(path, func(t *testing.T) {
			env := testApp.NewEnv(t)

			resp := env.Client.Post(t, path, body)

			require.Equal(t, http.StatusRequestEntityTooLarge, resp.StatusCode)
			respBody := DecodeBody[ErrorResponseBody](t, resp)
			assert.Equal(t, RequestBodyTooLargeCode, respBody.Error.Code)
		})
	}
}
