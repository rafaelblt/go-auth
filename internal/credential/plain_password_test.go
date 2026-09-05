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
		desc           string
		input          string
		expectedIssues validation.Issues
	}{
		{
			desc:           "empty input",
			input:          "",
			expectedIssues: validation.Issues{validation.IssueTooShort(PlainPasswordMinCodePoints, validation.UnitCodePoint)},
		},
		{
			desc:           "input too short",
			input:          strings.Repeat("a", PlainPasswordMinCodePoints-1),
			expectedIssues: validation.Issues{validation.IssueTooShort(PlainPasswordMinCodePoints, validation.UnitCodePoint)},
		},
		{
			desc:           "input too long",
			input:          strings.Repeat("a", PlainPasswordMaxBytes+1),
			expectedIssues: validation.Issues{validation.IssueTooLong(PlainPasswordMaxBytes, validation.UnitByte)},
		},
		{
			desc:  "valid input",
			input: "X8j5-30mWkPh",
		},
		{
			desc:  "with leading white space",
			input: "   mR&927Sa8.5f",
		},
		{
			desc:  "with traling white space",
			input: "9eF7}{d7[X$@   ",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			password, issues := NewPlainPassword(tC.input)

			if len(tC.expectedIssues) == 0 {
				require.Truef(t, issues.IsEmpty(), "unexpected issues: %s", issues)
				assert.Equal(t, tC.input, password.Value())
				return
			}

			require.Falsef(t, issues.IsEmpty(), "expected issues: %s", tC.expectedIssues)
			assert.True(t, password.IsZero())
			assert.ElementsMatch(t, tC.expectedIssues, issues)
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
