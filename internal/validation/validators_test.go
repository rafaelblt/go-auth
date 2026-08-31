package validation

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func issuer(code string) Validator[string] {
	return func(string) *Issue {
		return &Issue{code: code, details: map[string]any{}}
	}
}

func passing[T any]() Validator[T] {
	return func(T) *Issue { return nil }
}

func TestValidate(t *testing.T) {
	testCases := []struct {
		desc          string
		validators    []Validator[string]
		expectedCodes []string
	}{
		{
			desc:          "no validators",
			validators:    nil,
			expectedCodes: []string{},
		},
		{
			desc:          "all validators pass",
			validators:    []Validator[string]{passing[string](), passing[string]()},
			expectedCodes: []string{},
		},
		{
			desc:          "single validator fails",
			validators:    []Validator[string]{passing[string](), issuer("A")},
			expectedCodes: []string{"A"},
		},
		{
			desc:          "all validators fail keeping order",
			validators:    []Validator[string]{issuer("A"), issuer("B"), issuer("C")},
			expectedCodes: []string{"A", "B", "C"},
		},
		{
			desc:          "mixed results keeping order",
			validators:    []Validator[string]{issuer("A"), passing[string](), issuer("B")},
			expectedCodes: []string{"A", "B"},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			issues := Validate("any value", tC.validators...)

			require.NotNil(t, issues)
			codes := make([]string, len(issues))
			for i, iss := range issues {
				codes[i] = iss.Code()
			}
			assert.Equal(t, tC.expectedCodes, codes)
		})
	}
}

func TestValidate_PassesValueToEachValidator(t *testing.T) {
	received := make([]string, 0, 2)
	spy := func(value string) *Issue {
		received = append(received, value)
		return nil
	}

	Validate("some-value", spy, spy)

	assert.Equal(t, []string{"some-value", "some-value"}, received)
}

func TestValidate_WithNonStringType(t *testing.T) {
	code := "NOT_POSITIVE"
	positive := func(value int) *Issue {
		if value <= 0 {
			return &Issue{code: code, details: map[string]any{}}
		}
		return nil
	}

	assert.Empty(t, Validate(1, positive))

	iss := testutil.Only(t, Validate(0, positive))
	assert.Equal(t, code, iss.Code())
}

func TestLength(t *testing.T) {
	testCases := []struct {
		desc            string
		min             int
		max             int
		value           string
		expectedCode    string
		expectedDetails map[string]any
	}{
		{desc: "shorter than min", min: 3, max: 32, value: "ab", expectedCode: CodeTooShort, expectedDetails: map[string]any{"min": 3}},
		{desc: "empty value with positive min", min: 3, max: 32, value: "", expectedCode: CodeTooShort, expectedDetails: map[string]any{"min": 3}},
		{desc: "longer than max", min: 3, max: 5, value: "abcdef", expectedCode: CodeTooLong, expectedDetails: map[string]any{"max": 5}},
		{desc: "at min boundary", min: 3, max: 5, value: "abc"},
		{desc: "at max boundary", min: 3, max: 5, value: "abcde"},
		{desc: "between boundaries", min: 3, max: 5, value: "abcd"},
		{desc: "empty value with zero min", min: 0, max: 5, value: ""},
		{desc: "multibyte runes counted as runes", min: 3, max: 3, value: "ção"},
		{desc: "multibyte runes above max", min: 1, max: 2, value: "ção", expectedCode: CodeTooLong, expectedDetails: map[string]any{"max": 2}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := Length(tC.min, tC.max)(tC.value)

			if tC.expectedCode == "" {
				assert.Nil(t, iss)
				return
			}

			require.NotNil(t, iss)
			assert.Equal(t, tC.expectedCode, iss.Code())
			assert.Equal(t, tC.expectedDetails, iss.Details())
		})
	}
}

func TestAllowedChars(t *testing.T) {
	allowed := shared.NewSetFrom([]rune("abc123")...)

	testCases := []struct {
		desc    string
		value   string
		isValid bool
	}{
		{desc: "all chars allowed", value: "abc123", isValid: true},
		{desc: "empty value", value: "", isValid: true},
		{desc: "repeated allowed chars", value: "aaabbb", isValid: true},
		{desc: "single disallowed char", value: "abz", isValid: false},
		{desc: "disallowed char at start", value: "zabc", isValid: false},
		{desc: "uppercase is not allowed", value: "Abc", isValid: false},
		{desc: "whitespace is not allowed", value: "ab c", isValid: false},
		{desc: "multibyte char is not allowed", value: "abç", isValid: false},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := AllowedChars(allowed)(tC.value)

			if tC.isValid {
				assert.Nil(t, iss)
				return
			}

			require.NotNil(t, iss)
			assert.Equal(t, IssueInvalidChars(), *iss)
		})
	}
}

func TestAllowedChars_WithMultibyteAllowedSet(t *testing.T) {
	allowed := shared.NewSetFrom([]rune("çãô")...)
	validator := AllowedChars(allowed)

	assert.Nil(t, validator("çãô"))
	assert.NotNil(t, validator("çac"))
}

func TestAllowedChars_WithEmptySet(t *testing.T) {
	validator := AllowedChars(shared.NewSet[rune]())

	assert.Nil(t, validator(""))
	assert.NotNil(t, validator("a"))
}
