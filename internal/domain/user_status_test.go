package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

func TestIsActive_ShouldReturnTrue_WhenStatusIsActive(t *testing.T) {
	testCases := []struct {
		status domain.UserStatus
		expect bool
	}{
		{ status: domain.UserStatusActive, expect: true },
	}
	for _, tC := range testCases {
		t.Run(tC.status.String(), func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.status.IsActive())
		})
	}
}
