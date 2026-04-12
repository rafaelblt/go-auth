package infra_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnitOfWork_WithTxNil(t *testing.T) {
	uow, err := infra.NewUnitOfWork(nil)
	assert.Error(t, err)
	assert.Nil(t, uow)
}

func TestNewUnitOfWork_WithValidTx(t *testing.T) {
	tx := testDB.NewTx(t)

	uow, err := infra.NewUnitOfWork(tx)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestUnitOfWork_Do_ClosesTransaction(t *testing.T) {
	ctx := context.Background()

	tx := testDB.NewTx(t)
	uow, err := infra.NewUnitOfWork(tx)
	require.NoError(t, err)
	
	err = uow.Do(ctx, func(deps usecase.UowDeps) error { return nil })

	assert.NoError(t, err)
	assert.ErrorIs(t, tx.Rollback(ctx), pgx.ErrTxClosed, "the tx has not been closed")
}

func TestUnitOfWork_Do_WithUserWriter(t *testing.T) {
	ctx := context.Background()

	uow, err := infra.NewUnitOfWork(testDB.NewTx(t))
	require.NoError(t, err)

	user := testutil.DefaultUser(t)

	err = uow.Do(ctx, func(deps usecase.UowDeps) error {
		return deps.UserWriter.Save(ctx, user)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM users WHERE id=$1 )`
	assert.NoError(t, testDB.NewTx(t).QueryRow(ctx, query, user.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}

func TestUnitOfWork_Do_WithCredentialWriter(t *testing.T) {
	ctx := context.Background()

	uow, err := infra.NewUnitOfWork(testDB.NewTx(t))
	require.NoError(t, err)

	user := testutil.DefaultUser(t)
	credential := testutil.PasswordCredential(t, testutil.WithUserID(user.ID()))

	err = uow.Do(ctx, func(deps usecase.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, user))
		return deps.CredentialWriter.Save(ctx, credential)
	})

	assert.NoError(t, err)

	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM credentials WHERE id=$1 )`
	assert.NoError(t, testDB.NewTx(t).QueryRow(ctx, query, credential.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}
