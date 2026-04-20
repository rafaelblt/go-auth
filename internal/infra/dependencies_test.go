package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewDependencyContainer(t *testing.T) {
	db := testutil.NewDatabaseForTest(t, context.Background())
	testCases := []struct {
		desc      string
		cfg       infra.DependenciesConfig
		expectErr bool
	}{
		{
			desc:      "invalid config",
			cfg:       infra.DependenciesConfig{},
			expectErr: true,
		},
		{
			desc:      "valid config",
			cfg:       infra.DependenciesConfig{
				DatabaseConnection: db.ConnectionString(),
			},
			expectErr: false,
		},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			ctx := context.Background()
			container, err := infra.NewDependencyContainer(ctx, tC.cfg)
			if tC.expectErr {
				assert.Error(t, err)
				assert.Nil(t, container)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, container)
				container.Close()
			}
		})
	}
}

func TestDependencyContainer_BuildRegister(t *testing.T) {
	ctx := context.Background()
	db := testutil.NewDatabaseForTest(t, ctx)

	container, err := infra.NewDependencyContainer(ctx, infra.DependenciesConfig{
		DatabaseConnection: db.ConnectionString(),
	})
	require.NoError(t, err)
	defer container.Close()

	register, err := container.BuildRegister()

	require.NoError(t, err)
	require.NotZero(t, register)
}
