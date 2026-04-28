package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
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
	user := testutil.DefaultUser(t)

	err := repo.Save(context.Background(), user)

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
		user.ID().Value(),
		user.Username().String(),
		user.Status().String(),
		user.CreatedAt(),
		user.UpdatedAt(),
	).Scan(&exists)
	require.True(t, exists)
}

func TestUserRepo_ExistsByUsername_WhenUsernameExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Repo(db)
	user := testutil.DefaultUser(t)
	require.NoError(t, repo.Save(context.Background(), user))

	exists, err := repo.ExistsByUsername(context.Background(), user.Username())

	assert.NoError(t, err)
	assert.True(t, exists)
}

func TestUserRepo_ExistsByUsername_WhenUsernameNotExists(t *testing.T) {
	helper := NewUserRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Repo(db)
	user := testutil.DefaultUser(t)

	exists, err := repo.ExistsByUsername(context.Background(), user.Username())

	assert.NoError(t, err)
	assert.False(t, exists)
}
