package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
