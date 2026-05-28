package usecasetest

import "time"

type FakeClock struct {
	now time.Time
}

func NewFakeClock() *FakeClock {
	return &FakeClock{now: time.Now().UTC()}
}

func (c *FakeClock) Now() time.Time {
	return c.now
}

func (c *FakeClock) SetNow(now time.Time) {
	c.now = now
}
