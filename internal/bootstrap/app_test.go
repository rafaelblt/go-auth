package bootstrap

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

func appWithSessionCleanupForTest(cleanup *sessioncleanup.SessionCleanup) *App {
	return &App{deps: dependencies{UseCases: usecases{SessionCleanup: cleanup}}}
}

func taskNames(tasks []periodicTask) []string {
	names := []string{}
	for _, task := range tasks {
		names = append(names, task.name)
	}
	return names
}

// TESTS

// The intervals and timeouts are documented in
// docs/architecture/startup.md#background-tasks; a change here is a change
// there.
func TestBackgroundTasks_RunsTheKeyringRotationAndTheSessionCleanup(t *testing.T) {
	cleanup, err := sessioncleanup.New(sessioncleanup.Config{
		SessionCleaner: porttest.NewFakeSessionCleaner(),
		Clock:          porttest.NewFakeClock(),
		Retention:      time.Hour,
		BatchSize:      1,
		MaxBatches:     1,
	})
	require.NoError(t, err)

	tasks := appWithSessionCleanupForTest(cleanup).backgroundTasks()

	require.Equal(t, []string{"jwt_keyring_rotation", "session_cleanup"}, taskNames(tasks))
	assert.Equal(t, 10*time.Minute, tasks[0].interval)
	assert.Equal(t, 3*time.Second, tasks[0].timeout)
	assert.Equal(t, time.Hour, tasks[1].interval)
	assert.Equal(t, 5*time.Minute, tasks[1].timeout)
}

func TestBackgroundTasks_LeavesOutTheSessionCleanup_WhenItIsOff(t *testing.T) {
	tasks := appWithSessionCleanupForTest(nil).backgroundTasks()

	assert.Equal(t, []string{"jwt_keyring_rotation"}, taskNames(tasks))
}
