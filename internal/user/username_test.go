package user

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUsername(t *testing.T) {
	testCases := []struct {
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
			desc:        "empty value",
			input:       "",
			normalized:  "",
			expectedErr: []error{validation.IssueTooShort(UsernameMinLen)},
		},
		{
			desc:        "input too short",
			input:       strings.Repeat("a", UsernameMinLen-1),
			normalized:  "",
			expectedErr: []error{validation.IssueTooShort(UsernameMinLen)},
		},
		{
			desc:        "input too long",
			input:       strings.Repeat("a", UsernameMaxLen+1),
			normalized:  "",
			expectedErr: []error{validation.IssueTooLong(UsernameMaxLen)},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			username, err := NewUsername(tC.input)

			if len(tC.expectedErr) == 0 {
				require.NoError(t, err)
				assert.Equal(t, tC.normalized, username.String())
				return
			}

			require.Error(t, err)
			assert.True(t, username.IsZero())
			var issues validation.Issues
			require.ErrorAs(t, err, &issues)
			assert.ElementsMatch(t, tC.expectedErr, issues)
		})
	}
}
