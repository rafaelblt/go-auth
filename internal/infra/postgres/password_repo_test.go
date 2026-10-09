package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type PasswordRepoTestHelper struct {
	t  *testing.T
	db *pgxpool.Pool
}

func NewPasswordRepoTestHelper(t *testing.T) PasswordRepoTestHelper {
	db := poolFactory.AcquireWithMigrations(t)
	return PasswordRepoTestHelper{t, db}
}

func (helper PasswordRepoTestHelper) Repo() *postgres.PasswordRepo {
	helper.t.Helper()

	repo, err := postgres.NewPasswordRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper PasswordRepoTestHelper) TxRepo() (pgx.Tx, *postgres.PasswordRepo) {
	helper.t.Helper()

	tx, err := helper.db.Begin(helper.t.Context())
	require.NoError(helper.t, err)
	// t.Context() is already canceled when cleanups run.
	helper.t.Cleanup(func() { tx.Rollback(context.Background()) })

	repo, err := postgres.NewPasswordRepo(tx)
	require.NoError(helper.t, err)

	return tx, repo
}

func (helper PasswordRepoTestHelper) PersistentPassword() *password.Password {
	helper.t.Helper()

	usr := usertest.NewUser(helper.t, nil)
	helper.SaveUser(usr)

	pwd := passwordtest.NewPassword(helper.t, func(p *password.RestoreParams) {
		p.UserID = usr.ID()
	})
	require.NoError(helper.t, helper.Repo().Add(context.Background(), pwd))

	return pwd
}

// ChangedCopy returns a copy of pwd changed in memory to hash at changedAt, as
// held by a request that read the password and called ChangeHash.
func (helper PasswordRepoTestHelper) ChangedCopy(
	pwd *password.Password, hash string, changedAt time.Time,
) *password.Password {
	helper.t.Helper()

	cp := shared.ClonePtr(pwd)
	require.NoError(helper.t, cp.ChangeHash(passwordtest.MustHashed(helper.t, hash), changedAt))

	return cp
}

func (helper PasswordRepoTestHelper) SaveUser(usr *user.User) {
	helper.t.Helper()

	userRepo, err := postgres.NewUserRepo(helper.db)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userRepo.Add(context.Background(), usr))
}

func (helper PasswordRepoTestHelper) CheckPasswordIsSaved(pwd *password.Password) bool {
	var result bool
	err := helper.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM passwords WHERE
			id=$1 AND
			user_id=$2 AND
			hash=$3 AND
			created_at=$4 AND
			updated_at=$5
		)`,
		pwd.ID().Value(),
		pwd.UserID().Value(),
		pwd.Hash().Value(),
		pwd.CreatedAt(),
		pwd.UpdatedAt(),
	).Scan(&result)
	require.NoError(helper.t, err)
	return result
}

// TESTS

func TestPasswordRepo_Save(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	repo := helper.Repo()
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	pwd := passwordtest.NewPassword(t, func(params *password.RestoreParams) {
		params.UserID = usr.ID()
	})

	err := repo.Add(context.Background(), pwd)

	require.NoError(t, err)
	require.True(t, helper.CheckPasswordIsSaved(pwd))
}

func TestPasswordRepo_FindByID_ReturnsNil_WhenIDNotExists(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	pwd, err := repo.FindByID(ctx, password.NewID())

	assert.NoError(t, err)
	assert.Nil(t, pwd)
}

func TestPasswordRepo_FindByID_ReturnsPassword_WhenIDExists(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	pwd := passwordtest.NewPassword(t, func(params *password.RestoreParams) {
		params.UserID = usr.ID()
	})

	repo := helper.Repo()
	ctx := context.Background()
	require.NoError(t, repo.Add(ctx, pwd))

	found, err := repo.FindByID(ctx, pwd.ID())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, pwd.ID(), found.ID())
}

func TestPasswordRepo_FindByUserID_ReturnsNil_WhenNotExists(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	pwd, err := repo.FindByUserID(ctx, user.NewID())

	assert.NoError(t, err)
	assert.Nil(t, pwd)
}

func TestPasswordRepo_FindByUserID_ReturnsPassword_WhenExists(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	testPwd := passwordtest.NewPassword(t, func(p *password.RestoreParams) {
		p.UserID = usr.ID()
	})

	repo := helper.Repo()
	ctx := context.Background()
	require.NoError(t, repo.Add(ctx, testPwd))

	found, err := repo.FindByUserID(ctx, testPwd.UserID())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testPwd.ID(), found.ID())
}

func TestPasswordRepo_UpdateHash(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	pwd := helper.PersistentPassword()
	changed := helper.ChangedCopy(pwd, "new hash", time.Now().UTC())

	repo := helper.Repo()
	err := repo.UpdateHash(t.Context(), changed, pwd.Hash())

	assert.NoError(t, err)
	assert.True(t, helper.CheckPasswordIsSaved(changed))
}

func TestPasswordRepo_UpdateHash_ReturnsHashChanged_WhenStoredHashIsNotPrevious(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	pwd := helper.PersistentPassword()
	now := time.Now().UTC()
	first := helper.ChangedCopy(pwd, "first hash", now)
	second := helper.ChangedCopy(pwd, "second hash", now.Add(time.Second))

	repo := helper.Repo()
	require.NoError(t, repo.UpdateHash(t.Context(), first, pwd.Hash()))
	err := repo.UpdateHash(t.Context(), second, pwd.Hash())

	assert.ErrorIs(t, err, password.ErrHashChanged)
	assert.True(t, helper.CheckPasswordIsSaved(first), "first change was overwritten")
}

func TestPasswordRepo_UpdateHash_AppliesOnlyOneOfConcurrentChanges(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	pwd := helper.PersistentPassword()
	now := time.Now().UTC()
	first := helper.ChangedCopy(pwd, "first hash", now)
	second := helper.ChangedCopy(pwd, "second hash", now.Add(time.Second))
	tx1, repo1 := helper.TxRepo()
	tx2, repo2 := helper.TxRepo()

	require.NoError(t, repo1.UpdateHash(t.Context(), first, pwd.Hash()))
	result := make(chan error, 1)
	go func() { result <- repo2.UpdateHash(t.Context(), second, pwd.Hash()) }()
	requireWaitingOnLock(t, helper.db, tx2)
	require.NoError(t, tx1.Commit(t.Context()))

	assert.ErrorIs(t, <-result, password.ErrHashChanged)
	require.NoError(t, tx2.Commit(t.Context()))
	assert.True(t, helper.CheckPasswordIsSaved(first), "first change was overwritten")
}

func TestPasswordRepo_UpdateHash_Fails_WhenPasswordNotExists(t *testing.T) {
	helper := NewPasswordRepoTestHelper(t)
	pwd := passwordtest.NewPassword(t, nil)
	changed := helper.ChangedCopy(pwd, "new hash", time.Now().UTC())

	repo := helper.Repo()
	err := repo.UpdateHash(t.Context(), changed, pwd.Hash())

	require.Error(t, err)
	assert.NotErrorIs(t, err, password.ErrHashChanged)
}
