// Package infra holds SystemClock, the port.Clock every component is given
// outside tests, so that tests control time.
//
// See docs/architecture/conventions.md#clock.
package infra

import "time"

type SystemClock struct{}

func NewSystemClock() *SystemClock {
	return &SystemClock{}
}

func (c SystemClock) Now() time.Time {
	return time.Now().UTC()
}
