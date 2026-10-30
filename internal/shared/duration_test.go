package shared

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestShortDuration(t *testing.T) {
	testCases := []struct {
		duration time.Duration
		expected string
	}{
		{0, "0s"},
		{45 * time.Second, "45s"},
		{time.Minute, "1m"},
		{90 * time.Second, "1m30s"},
		{time.Hour, "1h"},
		{24 * time.Hour, "24h"},
		{90 * time.Minute, "1h30m"},
		{time.Hour + time.Second, "1h0m1s"},
		{-time.Hour, "-1h"},
		{1500 * time.Millisecond, "1.5s"},
	}
	for _, tC := range testCases {
		t.Run(tC.expected, func(t *testing.T) {
			assert.Equal(t, tC.expected, ShortDuration(tC.duration))
		})
	}
}
