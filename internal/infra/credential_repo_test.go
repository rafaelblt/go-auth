package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil/domaintest"
	"github.com/stretchr/testify/require"
)

// HELPER

type CredentialRepoTestHelper struct {
	t  *testing.T
}

func NewCredentialRepoTestHelper(t *testing.T) CredentialRepoTestHelper {
	return CredentialRepoTestHelper{t}
}

func (helper CredentialRepoTestHelper) DB() infra.PGDB {
	return dbProvider.NewPool(helper.t)
}

func (helper CredentialRepoTestHelper) Writer(db infra.PGDB) infra.CredentialRepo {
	helper.t.Helper()

	writer, err := infra.NewCredentialRepo(db)
	require.NoError(helper.t, err)

	return writer
}

func (helper CredentialRepoTestHelper) PersistentUser(db infra.PGDB) *domain.User {
	helper.t.Helper()
	require.NotNil(helper.t, db)

	user := domaintest.DefaultUser(helper.t)

	userRepo, err := infra.NewUserRepo(db)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userRepo.Save(context.Background(), user))

	return user
}

// TESTS

func TestCredentialRepo_Save(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	db := helper.DB()
	repo := helper.Writer(db)
	persistentUser := helper.PersistentUser(db)

	credential := domaintest.PasswordCredential(t,
		domaintest.WithUserID(persistentUser.ID()),
	)

	err := repo.Save(context.Background(), credential)

	require.NoError(t, err)
	var exists bool
	err = db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM credentials WHERE
			id=$1 AND
			user_id=$2 AND
			kind=$3 AND
			provider=$4 AND
			secret=$5 AND
			created_at=$6 AND
			updated_at=$7
		)`,
		credential.ID().Value(),
		credential.UserID().Value(),
		credential.Kind().String(),
		credential.Provider().String(),
		credential.Secret().Value(),
		credential.CreatedAt(),
		credential.UpdatedAt(),
	).Scan(&exists)
	require.True(t, exists)
}
