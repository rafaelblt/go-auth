package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUserStatus(t *testing.T) {
	testCases := []struct {
		input    string
		expected *domain.UserStatus
	}{
		{input: "active", expected: shared.Ptr(domain.UserStatusActive)},
		{input: "ACTIVE", expected: shared.Ptr(domain.UserStatusActive)},
		{input: "AcTiVe", expected: shared.Ptr(domain.UserStatusActive)},
		{input: " active ", expected: shared.Ptr(domain.UserStatusActive)},
		{input: "x", expected: nil},
	}
	for _, tC := range testCases {
		t.Run(tC.input, func(t *testing.T) {
			retrieved, err := domain.NewUserStatus(tC.input)
			if tC.expected != nil {
				require.NoError(t, err)
				assert.Equal(t, *tC.expected, retrieved)
			} else {
				assert.Error(t, err)
				assert.Zero(t, retrieved)
			}
		})
	}
}

func TestUserStatusString(t *testing.T) {
	testCases := []struct {
		status   domain.UserStatus
		expected string
	}{
		{status: domain.UserStatusActive, expected: "active"},
	}
	for _, tC := range testCases {
		t.Run(tC.expected, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.status.String())
		})
	}
}

func TestUserStatusIsActive(t *testing.T) {
	testCases := []struct {
		status domain.UserStatus
		expect bool
	}{
		{status: domain.UserStatusActive, expect: true},
	}
	for _, tC := range testCases {
		t.Run(tC.status.String(), func(t *testing.T) {
			assert.Equal(t, tC.expect, tC.status.IsActive())
		})
	}
}
