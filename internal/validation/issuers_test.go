package validation

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIssueTooLong(t *testing.T) {
	testCases := []struct {
		desc            string
		max             int
		expectedDetails map[string]any
	}{
		{desc: "positive max", max: 32, expectedDetails: map[string]any{"max": 32}},
		{desc: "zero max", max: 0, expectedDetails: map[string]any{"max": 0}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := IssueTooLong(tC.max)

			assert.Equal(t, CodeTooLong, iss.Code())
			assert.Equal(t, tC.expectedDetails, iss.Details())
		})
	}
}

func TestIssueTooShort(t *testing.T) {
	testCases := []struct {
		desc            string
		min             int
		expectedDetails map[string]any
	}{
		{desc: "positive min", min: 3, expectedDetails: map[string]any{"min": 3}},
		{desc: "zero min", min: 0, expectedDetails: map[string]any{"min": 0}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			iss := IssueTooShort(tC.min)

			assert.Equal(t, CodeTooShort, iss.Code())
			assert.Equal(t, tC.expectedDetails, iss.Details())
		})
	}
}

func TestIssuers_AreEqualForSameArgs(t *testing.T) {
	assert.Equal(t, IssueTooLong(32), IssueTooLong(32))
	assert.Equal(t, IssueTooShort(3), IssueTooShort(3))
	assert.NotEqual(t, IssueTooLong(32), IssueTooLong(10))
	assert.NotEqual(t, IssueTooLong(3), IssueTooShort(3))
}
