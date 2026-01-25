package user

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsActive_ShouldReturnTrue_WhenIsStatusActive(t *testing.T) {
	testCases := []struct {
		status Status
		expect bool
	}{
		{ status: StatusActive, expect: true },
	}
	for _, tC := range testCases {
		t.Run(tC.status.String(), func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.status.IsActive())
		})
	}
}
