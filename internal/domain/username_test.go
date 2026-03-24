package domain_test

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var usernamesData = []struct {
	desc        string
	input       string
	normalized  string
	expectedErr []error
}{
	{
		desc:        "valid case",
		input:       "blatantss",
		normalized:  "blatantss",
		expectedErr: []error{},
	},
	{
		desc:        "input with uppercase",
		input:       "BlatantSS",
		normalized:  "blatantss",
		expectedErr: []error{},
	},
	{
		desc:        "input with leading white space",
		input:       "  spiderman",
		normalized:  "spiderman",
		expectedErr: []error{},
	},
	{
		desc:        "input with trailing white space",
		input:       "venom  ",
		normalized:  "venom",
		expectedErr: []error{},
	},
	{
		desc:        "empty input",
		input:       "",
		normalized:  "",
		expectedErr: []error{domain.ErrUsernameTooShort},
	},
	{
		desc:        "input too short",
		input:       strings.Repeat("a", domain.UsernameMinLen-1),
		normalized:  "",
		expectedErr: []error{domain.ErrUsernameTooShort},
	},
	{
		desc:        "input too long",
		input:       strings.Repeat("a", domain.UsernameMaxLen+1),
		normalized:  "",
		expectedErr: []error{domain.ErrUsernameTooLong},
	},
}

func TestNewUsername(t *testing.T) {
	for _, tC := range usernamesData {
		t.Run(tC.desc, func(t *testing.T) {
			username, err := domain.NewUsername(tC.input)
			if tC.expectedErr == nil || len(tC.expectedErr) == 0 {
				require.NoError(t, err)
				assert.Equal(t, tC.normalized, username.String())
			} else {
				require.Error(t, err)
				assert.Empty(t, username)
				assert.Contains(t, tC.expectedErr, err)
			}
		})
	}
}

func TestValidateUsername_ShouldReturnExpectedErrors(t *testing.T) {
	for _, tC := range usernamesData {
		t.Run(tC.desc, func(t *testing.T) {
			errs := domain.ValidateUsername(tC.input)
			if tC.expectedErr == nil || len(tC.expectedErr) == 0 {
				assert.Empty(t, errs)
			} else {
				assert.Equal(t, tC.expectedErr, errs)
			}
		})
	}
}
