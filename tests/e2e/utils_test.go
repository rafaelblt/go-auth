package e2e

import (
	"net/http"
	"testing"

	"github.com/rafaelblt/go-auth/internal/testutil"
)

func DecodeBody[T any](t *testing.T, resp *http.Response) T {
	t.Helper()
	return testutil.DecodeJSON[T](t, resp.Body)
}
