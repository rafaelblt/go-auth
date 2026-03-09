package domain_test

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

var validPlainPasswords = []string{
	"12345678",
}

func TestNewPlainPassword_ShouldReturnPlainPassword_WhenInputIsValid(t *testing.T) {
	for _, input := range validPlainPasswords {
		t.Run(input, func(t *testing.T) {
			username, err := domain.NewPlainPassword(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, username)
			assert.Equal(t, input, username.String())
		})
	}
}

func TestNewPlainPassword_ShouldReturnError_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected error
	}{
		{
			desc:     "password empty",
			input:    "",
			expected: domain.ErrPlainPasswordTooShort,
		},
		{
			desc:     "password too short",
			input:    "123",
			expected: domain.ErrPlainPasswordTooShort,
		},
		{
			desc:     "password too long",
			input:    strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			expected: domain.ErrPlainPasswordTooLong,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			username, err := domain.NewPlainPassword(tC.input)
			assert.Zero(t, username)
			assert.ErrorIs(t, err, tC.expected)
		})
	}
}

func TestValidatePlainPassword_ShouldReturnExpectedErrors(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []error
	}{
		{
			desc:     "password empty",
			input:    "",
			expected: []error{domain.ErrPlainPasswordTooShort},
		},
		{
			desc:     "password too short",
			input:    "123",
			expected: []error{domain.ErrPlainPasswordTooShort},
		},
		{
			desc:     "password too long",
			input:    strings.Repeat("a", domain.PlainPasswordMaxLen+1),
			expected: []error{domain.ErrPlainPasswordTooLong},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := domain.ValidatePlainPassword(tC.input)
			assert.NotEmpty(t, errs)
			assert.Equal(t, tC.expected, errs)
		})
	}
}

func TestValidatePlainPassword_ShouldReturnEmpty_WhenInputIsValid(t *testing.T) {
	for _, input := range validPlainPasswords {
		t.Run(input, func(t *testing.T) {
			errs := domain.ValidatePlainPassword(input)
			assert.Empty(t, errs)
		})
	}
}
