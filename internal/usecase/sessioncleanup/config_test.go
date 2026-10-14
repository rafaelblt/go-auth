package sessioncleanup_test

import (
	"testing"

	"github.com/rafaelblt/go-auth/internal/usecase/sessioncleanup"
	"github.com/stretchr/testify/assert"
)

func TestNewSessionCleanup(t *testing.T) {
	testCases := []struct {
		desc      string
		mutate    func(cfg *sessioncleanup.Config)
		expectErr bool
	}{
		{
			desc:      "valid case",
			mutate:    func(cfg *sessioncleanup.Config) {},
			expectErr: false,
		},
		{
			desc:      "session cleaner nil",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.SessionCleaner = nil },
			expectErr: true,
		},
		{
			desc:      "clock nil",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.Clock = nil },
			expectErr: true,
		},
		{
			desc:      "retention zero",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.Retention = 0 },
			expectErr: true,
		},
		{
			desc:      "retention negative",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.Retention = -cfg.Retention },
			expectErr: true,
		},
		{
			desc:      "batch size zero",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.BatchSize = 0 },
			expectErr: true,
		},
		{
			desc:      "max batches zero",
			mutate:    func(cfg *sessioncleanup.Config) { cfg.MaxBatches = 0 },
			expectErr: true,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewTestHelper(t)
			cfg := helper.Config()
			tC.mutate(&cfg)

			uc, err := sessioncleanup.New(cfg)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Zero(t, uc)
				return
			}
			assert.NoError(t, err)
			assert.NotZero(t, uc)
		})
	}
}
