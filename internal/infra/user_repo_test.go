package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type UserRepoTestHelper struct {
	t *testing.T
}

func NewUserRepoTestHelper(t *testing.T) UserRepoTestHelper {
	return UserRepoTestHelper{t}
}

func (helper UserRepoTestHelper) DB() infra.PGDB {
	return dbProvider.NewPool(helper.t)
}

func (helper UserRepoTestHelper) Repo(db infra.PGDB) infra.UserRepo {
	helper.t.Helper()
	writer, err := infra.NewUserRepo(db)
	require.NoError(helper.t, err)
	return writer
}

func TestUserRepo_Save(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Repo(db)
	usr := usertest.NewUser(t, nil)

	err := repo.Save(context.Background(), usr)

	require.NoError(t, err)
	var exists bool
	err = db.QueryRow(context.Background(),
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
	).Scan(&exists)
	require.True(t, exists)
}

func TestUserRepo_ExistsByUsername_WhenUsernameExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Repo(db)
	usr := usertest.NewUser(t, nil)
	require.NoError(t, repo.Save(context.Background(), usr))

	exists, err := repo.ExistsByUsername(context.Background(), usr.Username())

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepo_ExistsByUsername_WhenUsernameNotExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Repo(db)
	usr := usertest.NewUser(t, nil)

	exists, err := repo.ExistsByUsername(context.Background(), usr.Username())

	assert.NoError(t, err)
	assert.False(t, exists)
}
