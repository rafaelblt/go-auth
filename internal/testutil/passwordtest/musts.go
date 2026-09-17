package passwordtest

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/stretchr/testify/require"
)

func MustPlain(t *testing.T, value string) password.Plain {
	t.Helper()
	plain, issues := password.NewPlain(value)
	require.Truef(t, issues.IsEmpty(), "%q is invalid to plain password: %s", value, issues)
	return plain
}

func MustHashed(t *testing.T, value string) password.Hashed {
	t.Helper()
	hashed, err := password.NewHashed(value)
	require.NoErrorf(t, err, "%q is invalid to hashed password", value)
	return hashed
}
