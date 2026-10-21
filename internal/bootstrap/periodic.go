package bootstrap

import (
	"context"
	"log/slog"
	"time"
)

type periodicTask struct {
	name     string
	interval time.Duration
	timeout  time.Duration
	// run returns the attributes its background task done line adds, as
	// key-value pairs.
	run func(ctx context.Context) ([]any, error)
}

func withoutAttrs(run func(ctx context.Context) error) func(ctx context.Context) ([]any, error) {
	return func(ctx context.Context) ([]any, error) {
		return nil, run(ctx)
	}
}

func runPeriodic(ctx context.Context, logger *slog.Logger, task periodicTask) {
	ticker := time.NewTicker(task.interval)
	defer ticker.Stop()

	logger = logger.With("task", task.name)
	logger.Info("background task started", "interval", task.interval)
	defer logger.Info("background task stopped")

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			execPeriodic(ctx, logger, task)
		}
	}
}

func execPeriodic(ctx context.Context, logger *slog.Logger, task periodicTask) {
	ctx, cancel := context.WithTimeout(ctx, task.timeout)
	defer cancel()

	// time.Now, not a port.Clock, for its monotonic reading. See
	// docs/architecture/conventions.md#clock.
	start := time.Now()
	attrs, err := task.run(ctx)
	if err != nil {
		logger.Error("background task failed", "error", err, "elapsed", time.Since(start))
		return
	}

	logger.Info("background task done", append([]any{"elapsed", time.Since(start)}, attrs...)...)
}
