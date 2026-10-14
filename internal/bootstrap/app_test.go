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
