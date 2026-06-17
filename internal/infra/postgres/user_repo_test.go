package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
}

func NewUserRepoTestHelper(t *testing.T) UserRepoTestHelper {
	db := poolFactory.Acquire(t)
	return UserRepoTestHelper{t, db}
}

func (helper UserRepoTestHelper) Repo() postgres.UserRepo {
	helper.t.Helper()
	repo, err := postgres.NewUserRepo(helper.db)
	require.NoError(helper.t, err)
	return repo
}

func (helper UserRepoTestHelper) CheckUserIsSaved(usr *user.User) bool {
	helper.t.Helper()
	var result bool
	err := helper.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM users WHERE
			id=$1 AND
			username=$2 AND
			status=$3 AND
			created_at=$4
			AND updated_at=$5
		)`,
		usr.ID().Value(),
		usr.Username().String(),
		usr.Status().String(),
		usr.CreatedAt(),
		usr.UpdatedAt(),
	).Scan(&result)
	require.NoError(helper.t, err)
	return result
}

func TestNewUserRepo_WithDBNil(t *testing.T) {
	uow, err := postgres.NewUserRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, uow)
}

func TestNewUserRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	uow, err := postgres.NewUserRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestUserRepo_Save(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()
	usr := usertest.NewUser(t, nil)

	err := repo.Save(context.Background(), usr)

	require.NoError(t, err)
	require.True(t, helper.CheckUserIsSaved(usr))
}

func TestUserRepo_ExistsByUsername_WhenUsernameExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()

	usr := usertest.NewUser(t, nil)
	require.NoError(t, repo.Save(context.Background(), usr))

	exists, err := repo.ExistsByUsername(context.Background(), usr.Username())

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepo_ExistsByUsername_WhenUsernameNotExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()

	usr := usertest.NewUser(t, nil)

	exists, err := repo.ExistsByUsername(context.Background(), usr.Username())

	assert.NoError(t, err)
	assert.False(t, exists)
}

func TestUserRepo_FindByUsername_ReturnsNil_WhenUsernameNotExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	usr, err := repo.FindByUsername(ctx, usertest.MustUsername(t, "not_exists"))

	assert.NoError(t, err)
	assert.Nil(t, usr)
}

func TestUserRepo_FindByUsername_ReturnsUser_WhenUsernameExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()
	testUser := usertest.NewUser(t, nil)
	
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, testUser))
	found, err := repo.FindByUsername(ctx, testUser.Username())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, found.ID(), found.ID())
}

func TestUserRepo_FindByID_ReturnsNil_WhenIDNotExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	usr, err := repo.FindByID(ctx, user.NewID())

	assert.NoError(t, err)
	assert.Nil(t, usr)
}

func TestUserRepo_FindByID_ReturnsUser_WhenIDExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	repo := helper.Repo()
	testUser := usertest.NewUser(t, nil)
	
	ctx := context.Background()
	require.NoError(t, repo.Save(ctx, testUser))
	found, err := repo.FindByID(ctx, testUser.ID())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, found.ID(), found.ID())
}
