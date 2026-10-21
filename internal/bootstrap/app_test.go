package bootstrap

import (
	"context"
	"slices"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func appWithSessionCleanupForTest(cleanup *sessioncleanup.SessionCleanup) *App {
	return &App{deps: dependencies{UseCases: usecases{SessionCleanup: cleanup}}}
}

// sessionCleanupForTest deletes at most 2 batches of 1 session.
func sessionCleanupForTest(t *testing.T, cleaner *porttest.FakeSessionCleaner) *sessioncleanup.SessionCleanup {
	t.Helper()

	cleanup, err := sessioncleanup.New(sessioncleanup.Config{
		SessionCleaner: cleaner,
		Clock:          porttest.NewFakeClock(),
		Retention:      time.Hour,
		BatchSize:      1,
		MaxBatches:     2,
	})
	require.NoError(t, err)
	return cleanup
}

func taskNamed(t *testing.T, tasks []periodicTask, name string) periodicTask {
	t.Helper()

	i := slices.IndexFunc(tasks, func(task periodicTask) bool { return task.name == name })
	require.NotEqual(t, -1, i, "no task named %s", name)
	return tasks[i]
}

func taskNames(tasks []periodicTask) []string {
	names := []string{}
	for _, task := range tasks {
		names = append(names, task.name)
	}
	return names
}

// TESTS

func TestBackgroundTasks_RunsTheKeyringRotationAndTheSessionCleanup(t *testing.T) {
	cleanup := sessionCleanupForTest(t, porttest.NewFakeSessionCleaner())

	tasks := appWithSessionCleanupForTest(cleanup).backgroundTasks()

	assert.Equal(t, []string{"jwt_keyring_rotation", "session_cleanup"}, taskNames(tasks))
	for _, task := range tasks {
		assert.Positive(t, task.interval, task.name)
		assert.Positive(t, task.timeout, task.name)
	}
}

func TestBackgroundTasks_LeavesOutTheSessionCleanup_WhenItIsOff(t *testing.T) {
	tasks := appWithSessionCleanupForTest(nil).backgroundTasks()

	assert.Equal(t, []string{"jwt_keyring_rotation"}, taskNames(tasks))
}

func TestBackgroundTasks_LogsWhatTheSessionCleanupDeleted(t *testing.T) {
	testCases := []struct {
		desc         string
		dead         int
		deleted      int
		limitReached bool
	}{
		{desc: "within its limit", dead: 1, deleted: 1, limitReached: false},
		{desc: "past its limit", dead: 3, deleted: 2, limitReached: true},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			cleaner := porttest.NewFakeSessionCleaner()
			cleaner.SetDead(tC.dead)
			tasks := appWithSessionCleanupForTest(sessionCleanupForTest(t, cleaner)).backgroundTasks()
			logger, buf := loggerWithLoggedLines()

			execPeriodic(context.Background(), logger, taskNamed(t, tasks, "session_cleanup"))

			line := testutil.Only(t, loggedLines(t, buf))
			assert.Equal(t, "background task done", line["msg"])
			assert.Equal(t, float64(tC.deleted), line["deleted"])
			assert.Equal(t, tC.limitReached, line["limit_reached"])
		})
	}
}
