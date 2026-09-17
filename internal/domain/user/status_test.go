package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStatus_String(t *testing.T) {
	testCases := []struct {
		status   Status
		expected string
	}{
		{status: StatusActive, expected: "active"},
	}
	for _, tC := range testCases {
		t.Run(tC.expected, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.status.String())
		})
	}
}

func TestStatus_IsActive(t *testing.T) {
	testCases := []struct {
		status Status
		expect bool
	}{
		{status: StatusActive, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.status.String(), func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.status.IsActive())
		})
	}
}

func TestStatus_IsZero(t *testing.T) {
	testCases := []struct {
		status Status
		expect bool
	}{
		{status: StatusActive, expect: false},
		{status: Status{}, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.status.String(), func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.status.IsZero())
		})
	}
}

func TestParseStatus(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected Status
		wantErr  bool
	}{
		{
			desc:     "active",
			value:    "active",
			expected: StatusActive,
		},
		{
			desc:     "uppercase active",
			value:    "ACTIVE",
			expected: StatusActive,
		},
		{
			desc:    "invalid value",
			value:   "invalid",
			wantErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			status, err := ParseStatus(tC.value)
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
