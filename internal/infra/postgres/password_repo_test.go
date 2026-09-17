package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type PasswordRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
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
