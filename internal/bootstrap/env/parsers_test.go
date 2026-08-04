package env

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringParser(t *testing.T) {
	value, err := StringParser("hello")

	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}

func TestIntParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected int
	}{
		{
			desc:     "positive",
			value:    "42",
			expected: 42,
		},
		{
			desc:     "negative",
			value:    "-7",
			expected: -7,
		},
		{
			desc:     "zero",
			value:    "0",
			expected: 0,
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := IntParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestIntParser_ReturnsError(t *testing.T) {
	value, err := IntParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}

func TestDurationParser(t *testing.T) {
	testCases := []struct {
		desc     string
		value    string
		expected time.Duration
	}{
		{
			desc:     "milliseconds",
			value:    "300ms",
			expected: 300 * time.Millisecond,
		},
		{
			desc:     "seconds",
			value:    "30s",
			expected: 30 * time.Second,
		},
		{
			desc:     "minutes",
			value:    "15m",
			expected: 15 * time.Minute,
		},
		{
			desc:     "hours",
			value:    "1h",
			expected: 1 * time.Hour,
		},
		{
			desc:     "minutes and seconds",
			value:    "12m30s",
			expected: (12 * time.Minute) + (30 * time.Second),
		},
		{
			desc:     "hours and minutes",
			value:    "2h45m",
			expected: (2 * time.Hour) + (45 * time.Minute),
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			value, err := DurationParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestDurationParser_ReturnsError(t *testing.T) {
	value, err := DurationParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}
