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
	run      func(ctx context.Context) error
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

	start := time.Now()
	if err := task.run(ctx); err != nil {
		logger.Error("background task failed", "err", err, "elapsed", time.Since(start))
		return
	}

	logger.Info("background task done", "elapsed", time.Since(start))
}
