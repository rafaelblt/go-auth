package domain_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
)

var validUsernames = []string{
	"blatantss",
}

func TestNewUsername_ShouldReturnUsername_WhenInputIsValid(t *testing.T) {
	for _, input := range validUsernames {
		t.Run(input, func(t *testing.T) {
			username, err := domain.NewUsername(input)
			assert.NoError(t, err)
			assert.NotEmpty(t, username)
			assert.Equal(t, input, username.String())
		})
	}
}

func TestNewUsername_ShouldReturnEmptyError_WhenInputIsEmpty(t *testing.T) {
	input := ""

	username, err := domain.NewUsername(input)

	assert.Empty(t, username)
	assert.Error(t, err)
	assert.EqualError(t, err, domain.ErrUsernameEmpty.Error())
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
			expected: []error{domain.ErrUsernameEmpty},
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
