package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
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
	cred := credentialtest.NewCredential(t, func(p *credential.RestoreParams) {
		p.UserID = usr.ID()
	})

	err = uow.Do(ctx, func(deps port.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, usr))
		return deps.CredentialWriter.Save(ctx, cred)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM credentials WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, cred.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}

func TestUnitOfWork_Do_WithSessionWriter(t *testing.T) {
	pool := poolFactory.Acquire(t)
	uow, err := postgres.NewUnitOfWork(pool)
	require.NoError(t, err)

	usr := usertest.NewUser(t, nil)
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})

	ctx := context.Background()
	err = uow.Do(ctx, func(deps port.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, usr))
		return deps.SessionWriter.Save(ctx, sess)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM sessions WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, sess.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}

func TestUnitOfWork_Do_WithRefreshTokenWriter(t *testing.T) {
	pool := poolFactory.Acquire(t)
	uow, err := postgres.NewUnitOfWork(pool)
	require.NoError(t, err)

	usr := usertest.NewUser(t, nil)
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
	})

	ctx := context.Background()
	err = uow.Do(ctx, func(deps port.UowDeps) error {
		require.NoError(t, deps.UserWriter.Save(ctx, usr))
		require.NoError(t, deps.SessionWriter.Save(ctx, sess))
		return deps.RefreshTokenWriter.Save(ctx, token)
	})

	assert.NoError(t, err)
	var exists bool
	query := `SELECT EXISTS( SELECT 1 FROM refresh_tokens WHERE id=$1 )`
	assert.NoError(t, pool.QueryRow(ctx, query, token.ID().Value()).Scan(&exists))
	assert.True(t, exists)
}
