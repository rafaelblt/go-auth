package sessioncleanup

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
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
	if cfg.Retention <= 0 {
		return nil, errors.New("retention zero or negative")
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
