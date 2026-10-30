package sessioncleanup

import (
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/shared"
)

type Config struct {
	SessionCleaner port.SessionCleaner
	Clock          port.Clock

	// Retention is how long a session is kept once it is dead.
	Retention time.Duration
	// BatchSize is how many sessions one delete takes at most, and MaxBatches
	// how many deletes one run makes at most.
	BatchSize  int
	MaxBatches int
}

func New(cfg Config) (*SessionCleanup, error) {
	if cfg.SessionCleaner == nil {
		return nil, errors.New("session cleaner nil")
	}
	if cfg.Clock == nil {
		return nil, errors.New("clock nil")
	}
	if err := CheckRetention(cfg.Retention); err != nil {
		return nil, fmt.Errorf("retention %w", err)
	}
	if cfg.BatchSize <= 0 {
		return nil, errors.New("batch size zero or negative")
	}
	if cfg.MaxBatches <= 0 {
		return nil, errors.New("max batches zero or negative")
	}
	uc := SessionCleanup{
		sessions:   cfg.SessionCleaner,
		clock:      cfg.Clock,
		retention:  cfg.Retention,
		batchSize:  cfg.BatchSize,
		maxBatches: cfg.MaxBatches,
	}
	return &uc, nil
}

// minRetention is far longer than any refresh runs, so the cleanup never
// deletes a session under a rotation. Reuse detection needs no margin.
//
// See docs/configuration.md#session_retention.
const minRetention = time.Hour

// CheckRetention is the rule for how long a dead session is kept, which
// SESSION_RETENTION sets when it is not off.
func CheckRetention(retention time.Duration) error {
	if retention < minRetention {
		return fmt.Errorf("must be at least %s, not %s", shared.ShortDuration(minRetention), shared.ShortDuration(retention))
	}
	return nil
}
