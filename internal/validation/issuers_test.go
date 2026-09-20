package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssueTooLong(t *testing.T) {
	testCases := []struct {
		desc            string
		max             int
		unit            LengthUnit
		expectedDetails map[string]any
	}{
		{
			desc: "positive max in code points",
			max:  32,
			unit: UnitCodePoint,
			expectedDetails: map[string]any{
				KeyMaxLength:  32,
				KeyUnitLength: UnitCodePoint,
			},
		},
		{
			desc: "positive max in bytes",
			max:  72,
			unit: UnitByte,
			expectedDetails: map[string]any{
				KeyMaxLength:  72,
				KeyUnitLength: UnitByte,
			},
		},
		{
			desc: "zero max",
			max:  0,
			unit: UnitCodePoint,
			expectedDetails: map[string]any{
				KeyMaxLength:  0,
				KeyUnitLength: UnitCodePoint,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := IssueTooLong(tC.max, tC.unit)

			assert.Equal(t, CodeTooLong, iss.Code())
			assert.Equal(t, tC.expectedDetails, iss.Details())
		})
	}
}

func TestIssueTooShort(t *testing.T) {
	testCases := []struct {
		desc            string
		min             int
		unit            LengthUnit
		expectedDetails map[string]any
	}{
		{
			desc: "positive min in code points",
			min:  3,
			unit: UnitCodePoint,
			expectedDetails: map[string]any{
				KeyMinLength:  3,
				KeyUnitLength: UnitCodePoint,
			},
		},
		{
			desc: "positive min in bytes",
			min:  4,
			unit: UnitByte,
			expectedDetails: map[string]any{
				KeyMinLength:  4,
				KeyUnitLength: UnitByte,
			},
		},
		{
			desc: "zero min",
			min:  0,
			unit: UnitCodePoint,
			expectedDetails: map[string]any{
				KeyMinLength:  0,
				KeyUnitLength: UnitCodePoint,
			},
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := IssueTooShort(tC.min, tC.unit)

			assert.Equal(t, CodeTooShort, iss.Code())
			assert.Equal(t, tC.expectedDetails, iss.Details())
		})
	}
}

func TestIssueInvalidChars(t *testing.T) {
	iss := IssueInvalidChars()

	assert.Equal(t, CodeInvalidCharacters, iss.Code())
	assert.Equal(t, map[string]any{}, iss.Details())
	assert.NotNil(t, iss.Details())
}

func TestIssuers_AreEqualForSameArgs(t *testing.T) {
	assert.Equal(t, IssueTooLong(32, UnitCodePoint), IssueTooLong(32, UnitCodePoint))
	assert.Equal(t, IssueTooShort(3, UnitCodePoint), IssueTooShort(3, UnitCodePoint))
	assert.NotEqual(t, IssueTooLong(32, UnitCodePoint), IssueTooLong(10, UnitCodePoint))
	assert.NotEqual(t, IssueTooLong(3, UnitCodePoint), IssueTooShort(3, UnitCodePoint))
	assert.Equal(t, IssueInvalidChars(), IssueInvalidChars())
	assert.NotEqual(t, IssueInvalidChars(), IssueTooShort(3, UnitCodePoint))
}

func TestIssuers_AreNotEqualForDifferentUnits(t *testing.T) {
	assert.NotEqual(t, IssueTooLong(32, UnitCodePoint), IssueTooLong(32, UnitByte))
	assert.NotEqual(t, IssueTooShort(3, UnitCodePoint), IssueTooShort(3, UnitByte))
}

func TestIssueRequired(t *testing.T) {
	iss := IssueRequired()

	assert.Equal(t, CodeRequired, iss.Code())
	assert.Empty(t, iss.Details())
}

func TestIssueNotPositive(t *testing.T) {
	iss := IssueNotPositive()

	assert.Equal(t, CodeNotPositive, iss.Code())
	assert.Empty(t, iss.Details())
}

func TestIssueNotAllowed(t *testing.T) {
	iss := IssueNotAllowed("json", "text")

	assert.Equal(t, CodeNotAllowed, iss.Code())
	assert.Equal(t, "json, text", iss.Details()[KeyAllowed])
}
