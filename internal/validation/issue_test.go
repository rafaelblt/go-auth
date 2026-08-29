package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssue_Code(t *testing.T) {
	testCases := []struct {
		desc     string
		issue    Issue
		expected string
	}{
		{
			desc:     "with code",
			issue:    Issue{code: "SOME_CODE", details: map[string]any{"key": "value"}},
			expected: "SOME_CODE",
		},
		{
			desc:     "zero value",
			issue:    Issue{},
			expected: "",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.issue.Code())
		})
	}
}

func TestIssue_Details(t *testing.T) {
	testCases := []struct {
		desc     string
		issue    Issue
		expected map[string]any
	}{
		{
			desc:     "with details",
			issue:    Issue{code: "SOME_CODE", details: map[string]any{"max": 10, "min": 2}},
			expected: map[string]any{"max": 10, "min": 2},
		},
		{
			desc:     "with empty details",
			issue:    Issue{code: "SOME_CODE", details: map[string]any{}},
			expected: map[string]any{},
		},
		{
			desc:     "zero value",
			issue:    Issue{},
			expected: nil,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.issue.Details())
		})
	}
}

func TestIssues_Error(t *testing.T) {
	testCases := []struct {
		desc     string
		issues   Issues
		expected string
	}{
		{
			desc:     "empty",
			issues:   Issues{},
			expected: "",
		},
		{
			desc:     "single issue",
			issues:   Issues{IssueTooShort(3)},
			expected: CodeTooShort,
		},
		{
			desc:     "multiple issues",
			issues:   Issues{IssueTooShort(3), IssueTooLong(32)},
			expected: CodeTooShort + "; " + CodeTooLong,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			var err error = tC.issues
			assert.Equal(t, tC.expected, err.Error())
		})
	}
}
