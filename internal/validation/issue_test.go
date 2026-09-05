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
			issue:    Issue{code: "SOME_CODE", details: map[string]any{KeyMaxLength: 10, KeyMinLength: 2}},
			expected: map[string]any{KeyMaxLength: 10, KeyMinLength: 2},
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

func TestIssues_String(t *testing.T) {
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
			issues:   Issues{IssueTooShort(3, UnitCodePoint)},
			expected: CodeTooShort,
		},
		{
			desc:     "multiple issues",
			issues:   Issues{IssueTooShort(3, UnitCodePoint), IssueTooLong(32, UnitCodePoint)},
			expected: CodeTooShort + "; " + CodeTooLong,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.issues.String())
		})
	}
}

func TestIssues_IsEmpty(t *testing.T) {
	testCases := []struct {
		desc     string
		issues   Issues
		expected bool
	}{
		{desc: "empty", issues: Issues{}, expected: true},
		{desc: "single issue", issues: Issues{IssueTooShort(3, UnitCodePoint)}, expected: false},
		{desc: "multiple issues", issues: Issues{IssueTooShort(3, UnitCodePoint), IssueTooLong(32, UnitCodePoint)}, expected: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			assert.Equal(t, tC.expected, tC.issues.IsEmpty())
		})
	}
}
