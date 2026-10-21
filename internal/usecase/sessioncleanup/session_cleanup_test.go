package sessioncleanup_test

import (
	"context"
	"errors"
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionCleanup_DeletesTheSessionsDeadBeforeNowMinusTheRetention(t *testing.T) {
	helper := NewTestHelper(t)
	helper.FakeSessionCleaner.SetDead(3)

	_, err := helper.UseCase().Execute(context.Background())

	require.NoError(t, err)
	assert.Zero(t, helper.FakeSessionCleaner.Dead())
	require.NotEmpty(t, helper.FakeSessionCleaner.Calls())
	for _, call := range helper.FakeSessionCleaner.Calls() {
		assert.Equal(t, helper.FakeClock.Now().Add(-helper.Retention), call.Cutoff)
	}
}

func TestSessionCleanup_DeletesInBatches_UntilOneComesBackShort(t *testing.T) {
	testCases := []struct {
		desc   string
		dead   int
		limits []int
	}{
		{desc: "nothing to delete", dead: 0, limits: []int{10}},
		{desc: "less than a batch", dead: 3, limits: []int{10}},
		{desc: "a batch and a half", dead: 15, limits: []int{10, 10}},
		// The second batch comes back empty, which is what shows the first one
		// left nothing.
		{desc: "exactly one batch", dead: 10, limits: []int{10, 10}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			helper.FakeSessionCleaner.SetDead(tC.dead)

			_, err := helper.UseCase().Execute(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tC.limits, helper.Limits())
			assert.Zero(t, helper.FakeSessionCleaner.Dead())
		})
	}
}

func TestSessionCleanup_LeavesTheRestToTheNextRun_AfterMaxBatches(t *testing.T) {
	helper := NewTestHelper(t)
	helper.FakeSessionCleaner.SetDead(100)

	_, err := helper.UseCase().Execute(context.Background())

	require.NoError(t, err)
	assert.Len(t, helper.FakeSessionCleaner.Calls(), helper.MaxBatches)
	assert.Equal(t, 100-helper.MaxBatches*helper.BatchSize, helper.FakeSessionCleaner.Dead())
}

func TestSessionCleanup_ReturnsHowManySessionsItDeleted_AndWhetherItReachedMaxBatches(t *testing.T) {
	testCases := []struct {
		desc     string
		dead     int
		expected sessioncleanup.Output
	}{
		{desc: "nothing to delete", dead: 0, expected: sessioncleanup.Output{Deleted: 0}},
		{desc: "a batch and a half", dead: 15, expected: sessioncleanup.Output{Deleted: 15}},
		{desc: "more than max batches", dead: 100, expected: sessioncleanup.Output{Deleted: 50, LimitReached: true}},
		// Every batch came back full, so the run cannot tell that nothing is
		// left.
		{desc: "exactly max batches", dead: 50, expected: sessioncleanup.Output{Deleted: 50, LimitReached: true}},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			helper.FakeSessionCleaner.SetDead(tC.dead)

			output, err := helper.UseCase().Execute(context.Background())

			require.NoError(t, err)
			assert.Equal(t, tC.expected, output)
		})
	}
}

func TestSessionCleanup_ReturnsTheError_WhenADeleteFails(t *testing.T) {
	helper := NewTestHelper(t)
	helper.FakeSessionCleaner.SetDead(100)
	expectedErr := errors.New("internal error")
	helper.FakeSessionCleaner.SetError(expectedErr)

	output, err := helper.UseCase().Execute(context.Background())

	assert.ErrorIs(t, err, expectedErr)
	assert.Zero(t, output)
	assert.Len(t, helper.FakeSessionCleaner.Calls(), 1, "kept deleting after a failure")
}
