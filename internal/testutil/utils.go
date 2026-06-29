package testutil

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Only[T any](t *testing.T, slice []T) T {
	t.Helper()
	require.Len(t, slice, 1)
	return slice[0]
}
