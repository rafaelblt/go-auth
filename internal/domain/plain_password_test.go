package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var validPlainPasswords = []string{
	"12345678",
}

func TestNewPlainPassword_ShouldReturnPlainPassword_WhenInputIsValid(t *testing.T) {
	for _, input := range validPlainPasswords {
		t.Run(input, func(t *testing.T) {
			username, err := NewPlainPassword(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, username)
			assert.Equal(t, input, username.String())
		})
	}
}

func TestNewPlainPassword_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	username, err := NewPlainPassword(input)

	assert.Empty(t, username)
	assert.Error(t, err)
	assert.EqualError(t, err, ErrPlainPasswordEmpty.Error())
}

func TestValidatePlainPassword_ShouldReturnExpectedErrors(t *testing.T) {
	testCases := []struct {
		desc     string
		input    string
		expected []DomainError
	}{
		{
			desc:     "password empty",
			input:    "",
			expected: []DomainError{ErrPlainPasswordEmpty},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			errs := ValidatePlainPassword(tC.input)
			assert.NotEmpty(t, errs)
			assert.Equal(t, tC.expected, errs)
		})
	}
}

func TestValidatePlainPassword_ShouldReturnEmpty_WhenInputIsValid(t *testing.T) {
	for _, input := range validPlainPasswords {
		t.Run(input, func(t *testing.T) {
			errs := ValidatePlainPassword(input)
			assert.Empty(t, errs)
		})
	}
}
