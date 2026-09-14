package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type RefreshTokenRepoTestHelper struct {
	t  *testing.T
	db *pgxpool.Pool
}

func NewRefreshTokenRepoTestHelper(t *testing.T) *RefreshTokenRepoTestHelper {
	db := poolFactory.AcquireWithMigrations(t)
	return &RefreshTokenRepoTestHelper{t, db}
}

func (helper *RefreshTokenRepoTestHelper) Repo() *postgres.RefreshTokenRepo {
	helper.t.Helper()

	repo, err := postgres.NewRefreshTokenRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper *RefreshTokenRepoTestHelper) TxRepo() (pgx.Tx, *postgres.RefreshTokenRepo) {
	helper.t.Helper()

	tx, err := helper.db.Begin(helper.t.Context())
	require.NoError(helper.t, err)
	// t.Context() is already canceled when cleanups run.
	helper.t.Cleanup(func() { tx.Rollback(context.Background()) })

	repo, err := postgres.NewRefreshTokenRepo(tx)
	require.NoError(helper.t, err)

	return tx, repo
}

func (helper *RefreshTokenRepoTestHelper) PersistentSession() *session.Session {
	helper.t.Helper()

	usr := usertest.NewUser(helper.t, nil)
	postgrestest.InsertUser(helper.t, helper.db, usr)

	sess := sessiontest.NewSession(helper.t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	postgrestest.InsertSession(helper.t, helper.db, sess)

	return sess
}

func (helper *RefreshTokenRepoTestHelper) PersistentRefreshToken() *session.RefreshToken {
	helper.t.Helper()

	sess := helper.PersistentSession()

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC().Add(time.Hour)
	})
	postgrestest.InsertRefreshToken(helper.t, helper.db, token)

	return token
}

// UsedCopy returns a copy of token spent in memory at usedAt, as held by a
// request that read the token and called Use.
func (helper *RefreshTokenRepoTestHelper) UsedCopy(token *session.RefreshToken, usedAt time.Time) *session.RefreshToken {
	helper.t.Helper()

	cp := shared.ClonePtr(token)
	require.NoError(helper.t, cp.Use(usedAt))

	return cp
}

// RequireWaitingOnLock waits until the connection of tx is blocked on a lock
// held by another transaction.
func (helper *RefreshTokenRepoTestHelper) RequireWaitingOnLock(tx pgx.Tx) {
	helper.t.Helper()

	pid := tx.Conn().PgConn().PID()
	sql := "SELECT wait_event_type = 'Lock' FROM pg_stat_activity WHERE pid = $1"

	require.Eventually(helper.t, func() bool {
		var waiting bool
		err := helper.db.QueryRow(helper.t.Context(), sql, pid).Scan(&waiting)
		return err == nil && waiting
	}, time.Second, 10*time.Millisecond, "transaction is not waiting on a lock")
}

func (helper *RefreshTokenRepoTestHelper) CheckRefreshTokenExists(token *session.RefreshToken) bool {
	helper.t.Helper()
	return postgrestest.CheckRefreshTokenExists(helper.t, helper.db, token)
}

// TESTS

func TestNewRefreshTokenRepo_WithDBNil(t *testing.T) {
	repo, err := postgres.NewRefreshTokenRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, repo)
}

func TestNewRefreshTokenRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	repo, err := postgres.NewRefreshTokenRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestRefreshTokenRepo_Add(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	sess := helper.PersistentSession()
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
	})

	repo := helper.Repo()
	err := repo.Add(t.Context(), token)

	assert.NoError(t, err)
	assert.True(t, helper.CheckRefreshTokenExists(token))
}

func TestRefreshTokenRepo_MarkUsed(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()
	require.NoError(t, token.Use(time.Now().UTC()))

	repo := helper.Repo()
	err := repo.MarkUsed(t.Context(), token)

	assert.NoError(t, err)
	assert.True(t, helper.CheckRefreshTokenExists(token))
}

func TestRefreshTokenRepo_MarkUsed_ReturnsTokenAlreadyUsed_WhenStoredTokenIsUsed(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()
	// Distinct uses, so an overwrite is visible at the column's precision.
	now := time.Now().UTC()
	first := helper.UsedCopy(token, now)
	second := helper.UsedCopy(token, now.Add(time.Second))

	repo := helper.Repo()
	require.NoError(t, repo.MarkUsed(t.Context(), first))
	err := repo.MarkUsed(t.Context(), second)

	assert.ErrorIs(t, err, session.ErrTokenAlreadyUsed)
	assert.True(t, helper.CheckRefreshTokenExists(first), "first use was overwritten")
}

func TestRefreshTokenRepo_MarkUsed_AppliesOnlyOneOfConcurrentUses(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()
	now := time.Now().UTC()
	first := helper.UsedCopy(token, now)
	second := helper.UsedCopy(token, now.Add(time.Second))
	tx1, repo1 := helper.TxRepo()
	tx2, repo2 := helper.TxRepo()

	require.NoError(t, repo1.MarkUsed(t.Context(), first))
	result := make(chan error, 1)
	go func() { result <- repo2.MarkUsed(t.Context(), second) }()
	helper.RequireWaitingOnLock(tx2)
	require.NoError(t, tx1.Commit(t.Context()))

	assert.ErrorIs(t, <-result, session.ErrTokenAlreadyUsed)
	require.NoError(t, tx2.Commit(t.Context()))
	assert.True(t, helper.CheckRefreshTokenExists(first), "first use was overwritten")
}

func TestRefreshTokenRepo_MarkUsed_Fails_WhenTokenNotUsed(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()

	repo := helper.Repo()
	err := repo.MarkUsed(t.Context(), token)

	require.Error(t, err)
	assert.NotErrorIs(t, err, session.ErrTokenAlreadyUsed)
}

func TestRefreshTokenRepo_MarkUsed_Fails_WhenTokenNotExists(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.UsedAt = shared.Ptr(time.Now().UTC())
	})

	repo := helper.Repo()
	err := repo.MarkUsed(t.Context(), token)

	require.Error(t, err)
	assert.NotErrorIs(t, err, session.ErrTokenAlreadyUsed)
}

func TestRefreshTokenRepo_FindByHash_ReturnsNil_WhenHashNotExists(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	repo := helper.Repo()

	hash := sessiontest.MustRefreshTokenHash(t, []byte{6, 7})
	found, err := repo.FindByHash(t.Context(), hash)

	assert.NoError(t, err)
	assert.Nil(t, found)
}

func TestRefreshTokenRepo_FindByHash_ReturnsToken_WhenHashExists(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()

	repo := helper.Repo()
	found, err := repo.FindByHash(t.Context(), token.Hash())

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, token.ID(), found.ID())
	assert.True(t, helper.CheckRefreshTokenExists(found), "found token differs from stored")
}
