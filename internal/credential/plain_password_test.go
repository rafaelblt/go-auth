package credential

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlainPassword(t *testing.T) {
	var testCases = []struct {
		desc        string
		input       string
		normalized  string
		expectedErr validation.Issues
	}{
		{
			desc:        "empty input",
			input:       "",
			expectedErr: validation.Issues{validation.IssueTooShort(PlainPasswordMinLen)},
		},
		{
			desc:        "input too short",
			input:       strings.Repeat("a", PlainPasswordMinLen-1),
			expectedErr: validation.Issues{validation.IssueTooShort(PlainPasswordMinLen)},
		},
		{
			desc:        "input too long",
			input:       strings.Repeat("a", PlainPasswordMaxLen+1),
			expectedErr: validation.Issues{validation.IssueTooLong(PlainPasswordMaxLen)},
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
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			password, err := NewPlainPassword(tC.input)

			if len(tC.expectedErr) == 0 {
				require.NoError(t, err)
				assert.Equal(t, tC.normalized, password.Value())
				return
			}

			require.Error(t, err)
			assert.True(t, password.IsZero())
			var issues validation.Issues
			require.ErrorAs(t, err, &issues)
			assert.ElementsMatch(t, tC.expectedErr, issues)
		})
	}
}

func TestPlainPassword_IsZero(t *testing.T) {
	testCases := []struct {
		desc     string
		pwd      PlainPassword
		expected bool
	}{
		{desc: "zero", pwd: PlainPassword{}, expected: true},
		{desc: "not zero", pwd: PlainPassword{"value"}, expected: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.pwd.IsZero())
		})
	}
}
