package config

import "time"

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
