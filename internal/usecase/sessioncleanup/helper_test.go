package sessioncleanup_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/testutil/porttest"
	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/stretchr/testify/require"
)

type TestHelper struct {
	t                  *testing.T
	FakeSessionCleaner *porttest.FakeSessionCleaner
	FakeClock          *porttest.FakeClock
	Retention          time.Duration
	BatchSize          int
	MaxBatches         int
}

func NewTestHelper(t *testing.T) TestHelper {
	helper := TestHelper{
		t:                  t,
		FakeSessionCleaner: porttest.NewFakeSessionCleaner(),
		FakeClock:          porttest.NewFakeClock(),
		Retention:          30 * 24 * time.Hour,
		BatchSize:          10,
		MaxBatches:         5,
	}
	return helper
}

func (helper *TestHelper) Config() sessioncleanup.Config {
	return sessioncleanup.Config{
		SessionCleaner: helper.FakeSessionCleaner,
		Clock:          helper.FakeClock,
		Retention:      helper.Retention,
		BatchSize:      helper.BatchSize,
		MaxBatches:     helper.MaxBatches,
	}
}

func (helper *TestHelper) UseCase() *sessioncleanup.SessionCleanup {
	helper.t.Helper()
	uc, err := sessioncleanup.New(helper.Config())
	require.NoError(helper.t, err)
	return uc
}

// Limits returns the limit of each call the use case made.
func (helper *TestHelper) Limits() []int {
	limits := []int{}
	for _, call := range helper.FakeSessionCleaner.Calls() {
		limits = append(limits, call.Limit)
	}
	return limits
}
