package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnitOfWork_WithTxBeginnerNil(t *testing.T) {
	uow, err := postgres.NewUnitOfWork(nil)
	assert.Error(t, err)
	assert.Nil(t, uow)
}

func TestNewUnitOfWork_WithValidTxBeginner(t *testing.T) {
	beginner := poolFactory.Acquire(t)

	uow, err := postgres.NewUnitOfWork(beginner)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestUnitOfWork_Do_WithUserWriter(t *testing.T) {
	ctx := context.Background()
	pool := poolFactory.Acquire(t)
	uow, err := postgres.NewUnitOfWork(pool)
	require.NoError(t, err)

	usr := usertest.NewUser(t, nil)

	err = uow.Do(ctx, func(deps port.UowDeps) error {
		return deps.UserWriter.Save(ctx, usr)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM users WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, usr.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}

func TestUnitOfWork_Do_WithCredentialWriter(t *testing.T) {
	ctx := context.Background()
	pool := poolFactory.Acquire(t)
	uow, err := postgres.NewUnitOfWork(pool)
	require.NoError(t, err)

	usr := usertest.NewUser(t, nil)
	credential := credentialtest.PasswordCredential(t, credentialtest.WithUserID(usr.ID()))

	err = uow.Do(ctx, func(deps port.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, usr))
		return deps.CredentialWriter.Save(ctx, credential)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM credentials WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, credential.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}
