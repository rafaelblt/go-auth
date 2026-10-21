// Package sessioncleanup deletes the sessions that have been dead for longer
// than the retention, revoked or with every refresh token expired, and their
// refresh tokens with them. It runs on a timer, not for a request.
//
// See docs/architecture/usecases.md#session-cleanup.
package sessioncleanup

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
)

type SessionCleanup struct {
	sessions   port.SessionCleaner
	clock      port.Clock
	retention  time.Duration
	batchSize  int
	maxBatches int
}

type Output struct {
	// Deleted is how many sessions the run deleted.
	Deleted int
	// LimitReached is set when the run stopped after MaxBatches full batches
	// rather than at a short one, so it may have left dead sessions behind.
	LimitReached bool
}

// Execute deletes the sessions that died before now minus the retention, a
// batch at a time, until a batch comes back short or MaxBatches have run. What
// is left is the next run's, so a backlog never makes one run long.
func (uc *SessionCleanup) Execute(ctx context.Context) (Output, error) {
	cutoff := uc.clock.Now().Add(-uc.retention)

	output := Output{}
	for range uc.maxBatches {
		deleted, err := uc.sessions.DeleteDeadBefore(ctx, cutoff, uc.batchSize)
		if err != nil {
			return Output{}, fmt.Errorf("dead sessions delete failed: %w", err)
		}
		output.Deleted += deleted
		if deleted < uc.batchSize {
			return output, nil
		}
	}
	output.LimitReached = true
	return output, nil
}
