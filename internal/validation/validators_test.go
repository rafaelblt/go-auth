package validation

import (
	"testing"
	"time"

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
		unit     LengthUnit
		value    string
		expected *Issue
	}{
		{desc: "code points: shorter than min", min: 3, unit: UnitCodePoint, value: "ab", expected: shared.Ptr(IssueTooShort(3, UnitCodePoint))},
		{desc: "code points: empty value with positive min", min: 3, unit: UnitCodePoint, value: "", expected: shared.Ptr(IssueTooShort(3, UnitCodePoint))},
		{desc: "code points: at min boundary", min: 3, unit: UnitCodePoint, value: "abc"},
		{desc: "code points: above min", min: 3, unit: UnitCodePoint, value: "abcd"},
		{desc: "code points: empty value with zero min", min: 0, unit: UnitCodePoint, value: ""},
		{desc: "code points: multibyte runes counted as runes", min: 3, unit: UnitCodePoint, value: "ção"},
		{desc: "code points: multibyte runes below min", min: 4, unit: UnitCodePoint, value: "ção", expected: shared.Ptr(IssueTooShort(4, UnitCodePoint))},
		{desc: "bytes: shorter than min", min: 3, unit: UnitByte, value: "ab", expected: shared.Ptr(IssueTooShort(3, UnitByte))},
		{desc: "bytes: empty value with positive min", min: 3, unit: UnitByte, value: "", expected: shared.Ptr(IssueTooShort(3, UnitByte))},
		{desc: "bytes: at min boundary", min: 3, unit: UnitByte, value: "abc"},
		{desc: "bytes: above min", min: 3, unit: UnitByte, value: "abcd"},
		{desc: "bytes: empty value with zero min", min: 0, unit: UnitByte, value: ""},
		{desc: "bytes: multibyte runes counted as bytes", min: 5, unit: UnitByte, value: "ção"},
		{desc: "bytes: multibyte runes below min", min: 6, unit: UnitByte, value: "ção", expected: shared.Ptr(IssueTooShort(6, UnitByte))},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := MinLength(tC.min, tC.unit)(tC.value)

			assert.Equal(t, tC.expected, iss)
		})
	}
}

func TestMaxLength(t *testing.T) {
	testCases := []struct {
		desc     string
		max      int
		unit     LengthUnit
		value    string
		expected *Issue
	}{
		{desc: "code points: longer than max", max: 5, unit: UnitCodePoint, value: "abcdef", expected: shared.Ptr(IssueTooLong(5, UnitCodePoint))},
		{desc: "code points: at max boundary", max: 5, unit: UnitCodePoint, value: "abcde"},
		{desc: "code points: below max", max: 5, unit: UnitCodePoint, value: "abcd"},
		{desc: "code points: empty value", max: 5, unit: UnitCodePoint, value: ""},
		{desc: "code points: empty value with zero max", max: 0, unit: UnitCodePoint, value: ""},
		{desc: "code points: any value with zero max", max: 0, unit: UnitCodePoint, value: "a", expected: shared.Ptr(IssueTooLong(0, UnitCodePoint))},
		{desc: "code points: multibyte runes counted as runes", max: 3, unit: UnitCodePoint, value: "ção"},
		{desc: "code points: multibyte runes above max", max: 2, unit: UnitCodePoint, value: "ção", expected: shared.Ptr(IssueTooLong(2, UnitCodePoint))},
		{desc: "bytes: longer than max", max: 5, unit: UnitByte, value: "abcdef", expected: shared.Ptr(IssueTooLong(5, UnitByte))},
		{desc: "bytes: at max boundary", max: 5, unit: UnitByte, value: "abcde"},
		{desc: "bytes: below max", max: 5, unit: UnitByte, value: "abcd"},
		{desc: "bytes: empty value", max: 5, unit: UnitByte, value: ""},
		{desc: "bytes: empty value with zero max", max: 0, unit: UnitByte, value: ""},
		{desc: "bytes: any value with zero max", max: 0, unit: UnitByte, value: "a", expected: shared.Ptr(IssueTooLong(0, UnitByte))},
		{desc: "bytes: multibyte runes counted as bytes", max: 5, unit: UnitByte, value: "ção"},
		{desc: "bytes: multibyte runes above max", max: 4, unit: UnitByte, value: "ção", expected: shared.Ptr(IssueTooLong(4, UnitByte))},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := MaxLength(tC.max, tC.unit)(tC.value)

			assert.Equal(t, tC.expected, iss)
		})
	}
}

func TestLengthValidators_UnitsDisagreeOnMultibyte(t *testing.T) {
	value := "ção" // 3 code points, 5 bytes

	assert.Nil(t, MaxLength(3, UnitCodePoint)(value))
	assert.Equal(t, shared.Ptr(IssueTooLong(3, UnitByte)), MaxLength(3, UnitByte)(value))

	assert.Equal(t, shared.Ptr(IssueTooShort(5, UnitCodePoint)), MinLength(5, UnitCodePoint)(value))
	assert.Nil(t, MinLength(5, UnitByte)(value))
}

func TestLengthValidators_PanicOnUnknownUnitAtConstruction(t *testing.T) {
	unknown := LengthUnit("unknown")

	assert.Panics(t, func() { MinLength(3, unknown) })
	assert.Panics(t, func() { MaxLength(3, unknown) })
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

func TestRequired(t *testing.T) {
	t.Run("returns issue when value is zero", func(t *testing.T) {
		iss := Required[string]()("")

		require.NotNil(t, iss)
		assert.Equal(t, CodeRequired, iss.Code())
	})

	t.Run("returns nil when value is set", func(t *testing.T) {
		assert.Nil(t, Required[string]()("value"))
	})
}

func TestPositive(t *testing.T) {
	t.Run("returns issue when value is zero", func(t *testing.T) {
		iss := Positive[int]()(0)

		require.NotNil(t, iss)
		assert.Equal(t, CodeNotPositive, iss.Code())
	})

	t.Run("returns issue when value is negative", func(t *testing.T) {
		iss := Positive[time.Duration]()(-time.Minute)

		require.NotNil(t, iss)
		assert.Equal(t, CodeNotPositive, iss.Code())
	})

	t.Run("returns nil when value is positive", func(t *testing.T) {
		assert.Nil(t, Positive[int]()(1))
		assert.Nil(t, Positive[time.Duration]()(time.Minute))
	})
}
