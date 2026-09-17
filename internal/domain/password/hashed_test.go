package password

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHashed(t *testing.T) {
	testCases := []struct {
		desc    string
		value   string
		wantErr bool
	}{
		{desc: "empty value", value: "", wantErr: true},
		{desc: "valid value", value: "$2a$10$hash", wantErr: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			hashed, err := NewHashed(tC.value)
			if tC.wantErr {
				assert.Error(t, err)
				assert.Zero(t, hashed)
				return
			}
			require.NoError(t, err)
			require.NotZero(t, hashed)
			assert.Equal(t, tC.value, hashed.Value())
		})
	}
}

func TestHashed_IsZero(t *testing.T) {
	testCases := []struct {
		desc   string
		hashed Hashed
		expect bool
	}{
		{desc: "not zero", hashed: Hashed{"value"}, expect: false},
		{desc: "zero", hashed: Hashed{}, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.hashed.IsZero())
		})
	}
}
