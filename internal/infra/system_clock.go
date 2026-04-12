package infra

import "time"

type SystemClock struct {}

func NewSystemClock() SystemClock {
	return SystemClock{}
}

func (c SystemClock) UtcNow() time.Time {
	return time.Now().UTC()
}
