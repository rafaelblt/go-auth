package domain_test

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var validUsernames = []string{
	"blatantss",
}

func TestNewUsername_ShouldReturnUsername_WhenInputIsValid(t *testing.T) {
	for _, input := range validUsernames {
		t.Run(input, func(t *testing.T) {
			username, err := domain.NewUsername(input)
			require.NoError(t, err)
			require.NotZero(t, username)
			assert.Equal(t, input, username.String())
		})
	}
}

func TestNewUsername_ShouldReturnError_WhenInputIsInvalid(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected error
	}{
		{
			desc:     "username empty",
			input:    "",
			expected: domain.ErrUsernameTooShort,
		},
		{
			desc:     "username too short",
			input:    strings.Repeat("a", domain.UsernameMinLen-1),
			expected: domain.ErrUsernameTooShort,
		},
		{
			desc:     "username too long",
			input:    strings.Repeat("a", domain.UsernameMaxLen+1),
			expected: domain.ErrUsernameTooLong,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			username, err := domain.NewUsername(tC.input)
			assert.Zero(t, username)
			assert.ErrorIs(t, err, tC.expected)
		})
	}
}

func TestValidateUsername_ShouldReturnExpectedErrors(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []error
	}{
		{
			desc:     "username empty",
			input:    "",
			expected: []error{domain.ErrUsernameTooShort},
		},
		{
			desc:     "username too short",
			input:    strings.Repeat("a", domain.UsernameMinLen-1),
			expected: []error{domain.ErrUsernameTooShort},
		},
		{
			desc:     "username too long",
			input:    strings.Repeat("a", domain.UsernameMaxLen+1),
			expected: []error{domain.ErrUsernameTooLong},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := domain.ValidateUsername(tC.input)
			assert.NotEmpty(t, errs)
			assert.Equal(t, tC.expected, errs)
		})
	}
}

func TestValidateUsername_ShouldReturnEmpty_WhenInputIsValid(t *testing.T) {
	for _, input := range validUsernames {
		t.Run(input, func(t *testing.T) {
			errs := domain.ValidateUsername(input)
			assert.Empty(t, errs)
		})
	}
}
