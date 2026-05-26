package usertest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/require"
)

func MustUsername(t *testing.T, value string) user.Username {
	t.Helper()
	username, err := user.NewUsername(value)
	require.NoError(t, err)
	return username
}
