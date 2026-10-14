package porttest

import (
	"context"
	"errors"
	"time"
)

// FakeSessionCleaner holds a number of dead sessions, and each call deletes up
// to its limit of them.
type FakeSessionCleaner struct {
	dead  int
	calls []FakeSessionCleanerCall
	err   error
}

type FakeSessionCleanerCall struct {
	Cutoff time.Time
	Limit  int
}

func NewFakeSessionCleaner() *FakeSessionCleaner {
	return &FakeSessionCleaner{
		calls: []FakeSessionCleanerCall{},
	}
}

func (c *FakeSessionCleaner) DeleteDeadBefore(
	ctx context.Context, cutoff time.Time, limit int,
) (int, error) {
	c.calls = append(c.calls, FakeSessionCleanerCall{
		Cutoff: cutoff,
		Limit:  limit,
	})

	if c.err != nil {
		return 0, c.err
	}
	if limit <= 0 {
		return 0, errors.New("limit zero or negative")
	}

	deleted := min(limit, c.dead)
	c.dead -= deleted
	return deleted, nil
}

func (c *FakeSessionCleaner) SetDead(dead int) {
	c.dead = dead
}

// Dead returns how many dead sessions are left.
func (c *FakeSessionCleaner) Dead() int {
	return c.dead
}

func (c *FakeSessionCleaner) Calls() []FakeSessionCleanerCall {
	return c.calls
}

func (c *FakeSessionCleaner) SetError(err error) {
	c.err = err
}
