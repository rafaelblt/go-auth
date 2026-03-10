package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func collectErrors(err error) []error {
	if err == nil {
		return nil
	}

	var out []error

	var walk func(error)
	walk = func(e error) {
		if e == nil {
			return
		}

		if u, ok := e.(interface{ Unwrap() []error }); ok {
			for _, inner := range u.Unwrap() {
				walk(inner)
			}
			return
		}

		if u, ok := e.(interface{ Unwrap() error }); ok {
			walk(u.Unwrap())
			return
		}

		out = append(out, e)
	}

	walk(err)
	return out
}

func RequireErrors(t *testing.T, err error, expected ...error) {
	t.Helper()

	actual := collectErrors(err)

	require.Len(t, actual, len(expected))

	for _, e := range expected {
		require.ErrorIs(t, err, e)
	}
}
