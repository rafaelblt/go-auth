package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var validUsernames = []string{
	"blatantss",
}

func TestNewUsername_ShouldReturnUsername_WhenInputIsValid(t *testing.T) {
	for _, input := range validUsernames {
		t.Run(input, func(t *testing.T) {
			username, err := NewUsername(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, username)
			assert.Equal(t, input, username.String())
		})
	}
}

func TestNewUsername_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	username, err := NewUsername(input)

	assert.Empty(t, username)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrUsernameEmpty.Error())
}

func TestValidateUsername_ShouldReturnExpectedErrors(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []DomainError
	}{
		{
			desc:     "username empty",
			input:    "",
			expected: []DomainError{ErrUsernameEmpty},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := ValidateUsername(tC.input)
			assert.NotEmpty(t, errs)
			assert.Equal(t, tC.expected, errs)
		})
	}
}

func TestValidateUsername_ShouldReturnEmpty_WhenInputIsValid(t *testing.T) {
	for _, input := range validUsernames {
		t.Run(input, func(t *testing.T) {
			errs := ValidateUsername(input)
			assert.Empty(t, errs)
		})
	}
}
