// Package infra holds SystemClock, the only code in the service that calls
// time.Now. Everything else takes a port.Clock, so tests control time.
//
// See docs/architecture/conventions.md#injected-clock.
package infra

import "time"

type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c SystemClock) Now() time.Time {
	return time.Now().UTC()
}
