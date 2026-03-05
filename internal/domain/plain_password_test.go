package domain_test

import (
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

func TestNewPlainPassword_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	username, err := domain.NewPlainPassword(input)

	assert.Empty(t, username)
	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrPlainPasswordEmpty.Error())
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
			expected: []error{domain.ErrPlainPasswordEmpty},
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
