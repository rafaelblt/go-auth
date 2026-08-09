package testutil

import (
	"encoding/json"
	"io"
	"testing"

	"github.com/stretchr/testify/require"
)

func DecodeJSON[T any](t *testing.T, body any) T {
	t.Helper()

	require.NotNil(t, body, "body nil")

	var result T

	switch b := body.(type) {
	case []byte:
		require.NoError(t, json.Unmarshal(b, &result), "json unmarshal failed")
	case io.ReadCloser:
		defer b.Close()
		require.NoError(t, json.NewDecoder(b).Decode(&result), "json decode failed")
	default:
		t.Fatalf("unsupported body type %T", body)
	}

	return result
}
