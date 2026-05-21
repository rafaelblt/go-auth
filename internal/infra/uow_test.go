package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil/domaintest"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewUnitOfWork_WithTxBeginnerNil(t *testing.T) {
	uow, err := infra.NewUnitOfWork(nil)
	assert.Error(t, err)
	assert.Nil(t, uow)
}

func TestNewUnitOfWork_WithValidTxBeginner(t *testing.T) {
	beginner := dbProvider.NewPool(t)

	uow, err := infra.NewUnitOfWork(beginner)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestUnitOfWork_Do_WithUserWriter(t *testing.T) {
	ctx := context.Background()
	pool := dbProvider.NewPool(t)

	uow, err := infra.NewUnitOfWork(pool)
	require.NoError(t, err)

	user := domaintest.NewUser(t, nil)

	err = uow.Do(ctx, func(deps usecase.UowDeps) error {
		return deps.UserWriter.Save(ctx, user)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM users WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, user.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}

func TestUnitOfWork_Do_WithCredentialWriter(t *testing.T) {
	ctx := context.Background()
	pool := dbProvider.NewPool(t)

	uow, err := infra.NewUnitOfWork(pool)
	require.NoError(t, err)

	user := domaintest.NewUser(t, nil)
	credential := domaintest.PasswordCredential(t, domaintest.WithUserID(user.ID()))

	err = uow.Do(ctx, func(deps usecase.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, user))
		return deps.CredentialWriter.Save(ctx, credential)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM credentials WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, credential.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}
