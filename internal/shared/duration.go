package shared

import (
	"strings"
	"time"
)

// ShortDuration writes d as time.Duration.String does, without the zero units
// at its end: 24h rather than 24h0m0s, 1h30m rather than 1h30m0s.
func ShortDuration(d time.Duration) string {
	s := d.String()
	if strings.HasSuffix(s, "m0s") {
		s = strings.TrimSuffix(s, "0s")
	}
	if strings.HasSuffix(s, "h0m") {
		s = strings.TrimSuffix(s, "0m")
	}
	return s
}
