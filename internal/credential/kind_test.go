package credential

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKind_String(t *testing.T) {
	testCases := []struct {
		status   Kind
		expected string
	}{
		{status: KindPassword, expected: "password"},
	}
	for _, tC := range testCases {
		t.Run(tC.expected, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.status.String())
		})
	}
}

func TestKind_IsZero(t *testing.T) {
	testCases := []struct {
		desc string
		kind Kind
		expect bool
	}{
		{desc: "not zero", kind: KindPassword, expect: false},
		{desc: "zero", kind: Kind{}, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.kind.IsZero())
		})
	}
}

func TestParseKind(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected Kind
		wantErr  bool
	}{
		{
			desc:     "password",
			value:    "password",
			expected: KindPassword,
		},
		{
			desc:     "uppercase password",
			value:    "PASSWORD",
			expected: KindPassword,
		},
		{
			desc:    "invalid value",
			value:   "invalid",
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			status, err := ParseKind(tC.value)
			if tC.wantErr {
				assert.Error(t, err)
				assert.Zero(t, status)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tC.expected, status)
		})
	}
}
