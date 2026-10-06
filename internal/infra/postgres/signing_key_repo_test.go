package postgres_test

import (
	"context"
	"crypto/rand"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type SigningKeyRepoTestHelper struct {
	t  *testing.T
	db *pgxpool.Pool
}

func NewSigningKeyRepoTestHelper(t *testing.T) *SigningKeyRepoTestHelper {
	db := poolFactory.AcquireWithMigrations(t)
	return &SigningKeyRepoTestHelper{t, db}
}

func (helper *SigningKeyRepoTestHelper) Repo() *postgres.SigningKeyRepo {
	helper.t.Helper()

	repo, err := postgres.NewSigningKeyRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper *SigningKeyRepoTestHelper) TxRepo() (pgx.Tx, *postgres.SigningKeyRepo) {
	helper.t.Helper()

	tx, err := helper.db.Begin(helper.t.Context())
	require.NoError(helper.t, err)
	// t.Context() is already canceled when cleanups run.
	helper.t.Cleanup(func() { tx.Rollback(context.Background()) })

	repo, err := postgres.NewSigningKeyRepo(tx)
	require.NoError(helper.t, err)

	return tx, repo
}

func (helper *SigningKeyRepoTestHelper) RequireWaitingOnLock(tx pgx.Tx) {
	helper.t.Helper()
	requireWaitingOnLock(helper.t, helper.db, tx)
}

func (helper *SigningKeyRepoTestHelper) InsertSigningKey(key port.StoredSigningKey) {
	helper.t.Helper()
	postgrestest.InsertSigningKey(helper.t, helper.db, key)
}

func (helper *SigningKeyRepoTestHelper) CheckSigningKeyExists(key port.StoredSigningKey) bool {
	helper.t.Helper()
	return postgrestest.CheckSigningKeyExists(helper.t, helper.db, key)
}

func storedSigningKeyForTest(generation int64) port.StoredSigningKey {
	seed := make([]byte, 32)
	rand.Read(seed)
	now := time.Now().UTC().Truncate(time.Microsecond)
	return port.StoredSigningKey{
		Generation: generation,
		Seed:       seed,
		ActiveAt:   now.Add(time.Duration(generation) * time.Hour),
		CreatedAt:  now,
	}
}

// TESTS

func TestNewSigningKeyRepo_WithDBNil(t *testing.T) {
	repo, err := postgres.NewSigningKeyRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, repo)
}

func TestNewSigningKeyRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	repo, err := postgres.NewSigningKeyRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, repo)
}

func TestSigningKeyRepo_List_ReturnsNoKey_WhenTheTableIsEmpty(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)

	repo := helper.Repo()
	keys, err := repo.List(t.Context())

	assert.NoError(t, err)
	assert.Empty(t, keys)
}

func TestSigningKeyRepo_List_ReturnsTheKeys_ByGeneration(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)
	first := storedSigningKeyForTest(1)
	second := storedSigningKeyForTest(2)
	helper.InsertSigningKey(second)
	helper.InsertSigningKey(first)

	repo := helper.Repo()
	keys, err := repo.List(t.Context())

	require.NoError(t, err)
	assert.Equal(t, []port.StoredSigningKey{first, second}, keys)
}

func TestSigningKeyRepo_Add(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)
	key := storedSigningKeyForTest(1)

	repo := helper.Repo()
	err := repo.Add(t.Context(), key)

	assert.NoError(t, err)
	assert.True(t, helper.CheckSigningKeyExists(key))
}

func TestSigningKeyRepo_Add_KeepsTheStoredKey_WhenTheGenerationIsTaken(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)
	stored := storedSigningKeyForTest(1)
	rival := storedSigningKeyForTest(1)
	helper.InsertSigningKey(stored)

	repo := helper.Repo()
	err := repo.Add(t.Context(), rival)

	assert.NoError(t, err)
	assert.True(t, helper.CheckSigningKeyExists(stored), "stored key was overwritten")
	assert.False(t, helper.CheckSigningKeyExists(rival))
}

func TestSigningKeyRepo_Add_AppliesOnlyOneOfConcurrentAdds(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)
	first := storedSigningKeyForTest(1)
	second := storedSigningKeyForTest(1)
	tx1, repo1 := helper.TxRepo()
	tx2, repo2 := helper.TxRepo()

	require.NoError(t, repo1.Add(t.Context(), first))
	result := make(chan error, 1)
	go func() { result <- repo2.Add(t.Context(), second) }()
	helper.RequireWaitingOnLock(tx2)
	require.NoError(t, tx1.Commit(t.Context()))

	assert.NoError(t, <-result)
	require.NoError(t, tx2.Commit(t.Context()))
	assert.True(t, helper.CheckSigningKeyExists(first), "first add was overwritten")
	assert.False(t, helper.CheckSigningKeyExists(second))
}

func TestSigningKeyRepo_DeleteBefore_DeletesOnlyLowerGenerations(t *testing.T) {
	helper := NewSigningKeyRepoTestHelper(t)
	third := storedSigningKeyForTest(3)
	helper.InsertSigningKey(storedSigningKeyForTest(1))
	helper.InsertSigningKey(storedSigningKeyForTest(2))
	helper.InsertSigningKey(third)

	repo := helper.Repo()
	err := repo.DeleteBefore(t.Context(), 3)

	require.NoError(t, err)
	keys, err := repo.List(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []port.StoredSigningKey{third}, keys)

	err = repo.DeleteBefore(t.Context(), 1)

	require.NoError(t, err)
	keys, err = repo.List(t.Context())
	require.NoError(t, err)
	assert.Equal(t, []port.StoredSigningKey{third}, keys, "nothing below generation 1 to delete")
}
