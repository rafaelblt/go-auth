package domain_test

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var passwordCases = []struct {
	desc        string
	input       string
	normalized  string
	expectedErr []error
}{
	{
		desc:        "empty input",
		input:       "",
		expectedErr: []error{domain.ErrPlainPasswordTooShort},
	},
	{
		desc:        "input too short",
		input:       strings.Repeat("a", domain.PlainPasswordMinLen-1),
		expectedErr: []error{domain.ErrPlainPasswordTooShort},
	},
	{
		desc:        "input too long",
		input:       strings.Repeat("a", domain.PlainPasswordMaxLen+1),
		expectedErr: []error{domain.ErrPlainPasswordTooLong},
	},
	{
		desc:       "valid input",
		input:      "X8j5-30mWkPh",
		normalized: "X8j5-30mWkPh",
	},
	{
		desc:       "with leading white space",
		input:      "   mR&927Sa8.5f",
		normalized: "mR&927Sa8.5f",
	},
	{
		desc:       "with traling white space",
		input:      "9eF7}{d7[X$@   ",
		normalized: "9eF7}{d7[X$@",
	},
	
}

func TestNewPlainPassword(t *testing.T) {
	for _, tC := range passwordCases {
		t.Run(tC.desc, func(t *testing.T) {
			password, err := domain.NewPlainPassword(tC.input)
			if len(tC.expectedErr) == 0 {
				require.NoError(t, err)
				assert.Equal(t, tC.normalized, password.Value())
			} else {
				require.Error(t, err)
				assert.True(t, password.IsZero())
				var verr *domain.ValidationError
				require.ErrorAs(t, err, &verr)
				assert.Equal(t, tC.expectedErr, verr.Unwrap())
			}
		})
	}
}
