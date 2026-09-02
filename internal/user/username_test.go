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
		desc           string
		input          string
		normalized     string
		expectedIssues validation.Issues
	}{
		{
			desc:       "valid case",
			input:      "blatantss",
			normalized: "blatantss",
		},
		{
			desc:       "input with uppercase",
			input:      "BlatantSS",
			normalized: "blatantss",
		},
		{
			desc:           "empty value",
			input:          "",
			normalized:     "",
			expectedIssues: validation.Issues{validation.IssueTooShort(UsernameMinLen)},
		},
		{
			desc:           "input too short",
			input:          strings.Repeat("a", UsernameMinLen-1),
			normalized:     "",
			expectedIssues: validation.Issues{validation.IssueTooShort(UsernameMinLen)},
		},
		{
			desc:           "input too long",
			input:          strings.Repeat("a", UsernameMaxLen+1),
			normalized:     "",
			expectedIssues: validation.Issues{validation.IssueTooLong(UsernameMaxLen)},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			username, issues := NewUsername(tC.input)

			if len(tC.expectedIssues) == 0 {
				require.Truef(t, issues.IsEmpty(), "unexpected issues: %s", issues)
				assert.Equal(t, tC.normalized, username.String())
				return
			}

			require.Falsef(t, issues.IsEmpty(), "expected issues: %s", tC.expectedIssues)
			assert.True(t, username.IsZero())
			assert.ElementsMatch(t, tC.expectedIssues, issues)
		})
	}
}
