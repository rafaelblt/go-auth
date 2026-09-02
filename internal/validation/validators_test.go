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

func TestMinLength(t *testing.T) {
	testCases := []struct {
		desc     string
		min      int
		value    string
		expected *Issue
	}{
		{desc: "shorter than min", min: 3, value: "ab", expected: shared.Ptr(IssueTooShort(3))},
		{desc: "empty value with positive min", min: 3, value: "", expected: shared.Ptr(IssueTooShort(3))},
		{desc: "at min boundary", min: 3, value: "abc"},
		{desc: "above min", min: 3, value: "abcd"},
		{desc: "empty value with zero min", min: 0, value: ""},
		{desc: "multibyte runes counted as runes", min: 3, value: "ção"},
		{desc: "multibyte runes below min", min: 4, value: "ção", expected: shared.Ptr(IssueTooShort(4))},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := MinLength(tC.min)(tC.value)

			assert.Equal(t, tC.expected, iss)
		})
	}
}

func TestMaxLength(t *testing.T) {
	testCases := []struct {
		desc     string
		max      int
		value    string
		expected *Issue
	}{
		{desc: "longer than max", max: 5, value: "abcdef", expected: shared.Ptr(IssueTooLong(5))},
		{desc: "at max boundary", max: 5, value: "abcde"},
		{desc: "below max", max: 5, value: "abcd"},
		{desc: "empty value", max: 5, value: ""},
		{desc: "empty value with zero max", max: 0, value: ""},
		{desc: "any value with zero max", max: 0, value: "a", expected: shared.Ptr(IssueTooLong(0))},
		{desc: "multibyte runes counted as runes", max: 3, value: "ção"},
		{desc: "multibyte runes above max", max: 2, value: "ção", expected: shared.Ptr(IssueTooLong(2))},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := MaxLength(tC.max)(tC.value)

			assert.Equal(t, tC.expected, iss)
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
