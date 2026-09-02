package usertest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func MustUsername(t *testing.T, value string) user.Username {
	t.Helper()
	username, issues := user.NewUsername(value)
	require.Truef(t, issues.IsEmpty(), "%q is invalid to username: %s", value, issues)
	return username
}
