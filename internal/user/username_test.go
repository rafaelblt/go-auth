package user

import (
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
			expectedErr: []error{ErrUsernameEmpty},
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
			var verr *validation.ValidationError
			require.ErrorAs(t, err, &verr)
			assert.Equal(t, tC.expectedErr, verr.Errors())
		})
	}
}
