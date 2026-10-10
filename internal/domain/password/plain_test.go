package password

import (
	"strings"
	"testing"

	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPlain(t *testing.T) {
	var testCases = []struct {
		desc           string
		input          string
		expectedIssues validation.Issues
	}{
		{
			desc:           "empty input",
			input:          "",
			expectedIssues: validation.Issues{validation.IssueTooShort(PlainMinCodePoints, validation.UnitCodePoint)},
		},
		{
			desc:           "input too short",
			input:          strings.Repeat("a", PlainMinCodePoints-1),
			expectedIssues: validation.Issues{validation.IssueTooShort(PlainMinCodePoints, validation.UnitCodePoint)},
		},
		{
			desc:           "input too long",
			input:          strings.Repeat("a", PlainMaxBytes+1),
			expectedIssues: validation.Issues{validation.IssueTooLong(PlainMaxBytes, validation.UnitByte)},
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
			password, issues := NewPlain(tC.input)

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

func TestNewPlainReplacing(t *testing.T) {
	current := Plain{"X8j5-30mWkPh"}
	testCases := []struct {
		desc           string
		input          string
		expectedIssues validation.Issues
	}{
		{
			desc:           "the current password",
			input:          "X8j5-30mWkPh",
			expectedIssues: validation.Issues{validation.IssueSameAsCurrent()},
		},
		{
			desc:           "too short",
			input:          strings.Repeat("a", PlainMinCodePoints-1),
			expectedIssues: validation.Issues{validation.IssueTooShort(PlainMinCodePoints, validation.UnitCodePoint)},
		},
		{
			desc:  "another password",
			input: "   mR&927Sa8.5f",
		},
		{
			desc:  "the current password in another case",
			input: "x8J5-30MwKpH",
		},
		{
			desc:  "the current password with white space",
			input: "X8j5-30mWkPh ",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			password, issues := NewPlainReplacing(tC.input, current)

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

func TestPlain_IsZero(t *testing.T) {
	testCases := []struct {
		desc     string
		pwd      Plain
		expected bool
	}{
		{desc: "zero", pwd: Plain{}, expected: true},
		{desc: "not zero", pwd: Plain{"value"}, expected: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.pwd.IsZero())
		})
	}
}
