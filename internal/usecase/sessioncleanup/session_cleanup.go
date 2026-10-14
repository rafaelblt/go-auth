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

// Execute deletes the sessions that died before now minus the retention, a
// batch at a time, until a batch comes back short or MaxBatches have run. What
// is left is the next run's, so a backlog never makes one run long.
func (uc *SessionCleanup) Execute(ctx context.Context) error {
	cutoff := uc.clock.Now().Add(-uc.retention)

	for range uc.maxBatches {
		deleted, err := uc.sessions.DeleteDeadBefore(ctx, cutoff, uc.batchSize)
		if err != nil {
			return fmt.Errorf("dead sessions delete failed: %w", err)
		}
		if deleted < uc.batchSize {
			return nil
		}
	}
	return nil
}
