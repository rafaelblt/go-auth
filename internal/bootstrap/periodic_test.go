package bootstrap

import (
	"context"
	"errors"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
)

// HELPERS

func periodicTaskForTest(run func(ctx context.Context) ([]any, error)) periodicTask {
	return periodicTask{name: "test", interval: time.Hour, timeout: time.Second, run: run}
}

// TESTS

func TestExecPeriodic_LogsOnlyElapsed_WhenTheRunAddsNothing(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	task := periodicTaskForTest(withoutAttrs(func(context.Context) error { return nil }))

	execPeriodic(context.Background(), logger, task)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "INFO", line["level"])
	assert.Equal(t, "background task done", line["msg"])
	assert.ElementsMatch(t, []string{"time", "level", "msg", "elapsed"}, slices.Collect(maps.Keys(line)))
}

func TestExecPeriodic_AddsTheAttributesOfTheRun_ToItsDoneLine(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	task := periodicTaskForTest(func(context.Context) ([]any, error) {
		return []any{"deleted", 3}, nil
	})

	execPeriodic(context.Background(), logger, task)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "background task done", line["msg"])
	assert.Contains(t, line, "elapsed")
	assert.Equal(t, float64(3), line["deleted"])
}

func TestExecPeriodic_LogsTheError_WhenTheRunFails(t *testing.T) {
	logger, buf := loggerWithLoggedLines()
	task := periodicTaskForTest(withoutAttrs(func(context.Context) error {
		return errors.New("sync failed")
	}))

	execPeriodic(context.Background(), logger, task)

	line := testutil.Only(t, loggedLines(t, buf))
	assert.Equal(t, "ERROR", line["level"])
	assert.Equal(t, "background task failed", line["msg"])
	assert.Equal(t, "sync failed", line["error"])
	assert.Contains(t, line, "elapsed")
}
