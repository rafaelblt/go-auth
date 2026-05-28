package sessiontest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/stretchr/testify/require"
)

func MustAccessToken(t *testing.T, value string) session.AccessToken {
	token, err := session.NewAccessToken(value)
	require.NoError(t, err)
	return token
}
