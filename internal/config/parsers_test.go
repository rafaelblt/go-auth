package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStringEnvParser(t *testing.T) {
	value, err := stringEnvParser("hello")

	require.NoError(t, err)
	assert.Equal(t, "hello", value)
}

func TestIntEnvParser(t *testing.T) {
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
			value, err := intEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestIntEnvParser_ReturnsError(t *testing.T) {
	value, err := intEnvParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}

func TestDurationEnvParser(t *testing.T) {
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
			value, err := durationEnvParser(tC.value)

			require.NoError(t, err)
			assert.Equal(t, tC.expected, value)
		})
	}
}

func TestDurationEnvParser_ReturnsError(t *testing.T) {
	value, err := durationEnvParser("invalid")

	assert.Error(t, err)
	assert.Zero(t, value)
}
