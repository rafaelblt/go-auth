package infra

import "time"

type SystemClock struct {}

func (c SystemClock) UtcNow() time.Time {
	return time.Now().UTC()
}
