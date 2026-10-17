package config

import (
	"time"

	"github.com/rafaelblt/go-auth/internal/validation"
)

// SessionRetention is how long a session stays stored once it is dead, before
// the session cleanup deletes it, or off, which keeps every session. Off is not
// a duration, so no duration, 0 included, can turn the cleanup off.
type SessionRetention struct {
	off      bool
	duration time.Duration
}

func SessionRetentionOff() SessionRetention {
	return SessionRetention{off: true}
}

func SessionRetentionOf(d time.Duration) SessionRetention {
	return SessionRetention{duration: d}
}

// Duration returns the retention, and false when it is off.
func (r SessionRetention) Duration() (time.Duration, bool) {
	if r.off {
		return 0, false
	}
	return r.duration, true
}

// minSessionRetention is far longer than any refresh runs, so the cleanup
// never deletes a session under a rotation. Reuse detection needs no margin.
//
// See docs/configuration.md#session_retention.
const minSessionRetention = time.Hour

var sessionRetentionValidators = []validation.Validator[time.Duration]{
	validation.Min(minSessionRetention),
}
