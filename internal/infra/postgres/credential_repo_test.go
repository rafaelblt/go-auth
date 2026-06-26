package postgres_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil/credentialtest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type CredentialRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
}

func NewCredentialRepoTestHelper(t *testing.T) CredentialRepoTestHelper {
	db := poolFactory.Acquire(t)
	return CredentialRepoTestHelper{t, db}
}

func (helper CredentialRepoTestHelper) Repo() *postgres.CredentialRepo {
	helper.t.Helper()

	repo, err := postgres.NewCredentialRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper CredentialRepoTestHelper) SaveUser(usr *user.User) {
	helper.t.Helper()

	userRepo, err := postgres.NewUserRepo(helper.db)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userRepo.Add(context.Background(), usr))
}

func (helper CredentialRepoTestHelper) CheckCredentialIsSaved(cred *credential.Credential) bool {
	var result bool
	err := helper.db.QueryRow(context.Background(),
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
		cred.ID().Value(),
		cred.UserID().Value(),
		cred.Kind().String(),
		cred.Provider().String(),
		cred.Secret().Value(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	).Scan(&result)
	require.NoError(helper.t, err)
	return result
}

// TESTS

func TestCredentialRepo_Save(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	repo := helper.Repo()
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	cred := credentialtest.NewCredential(t, func(params *credential.RestoreParams) {
		params.UserID = usr.ID()
	})

	err := repo.Add(context.Background(), cred)

	require.NoError(t, err)
	require.True(t, helper.CheckCredentialIsSaved(cred))
}

func TestCredentialRepo_FindByID_ReturnsNil_WhenIDNotExists(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	cred, err := repo.FindByID(ctx, credential.NewID())

	assert.NoError(t, err)
	assert.Nil(t, cred)
}

func TestCredentialRepo_FindByID_ReturnsCredential_WhenIDExists(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	cred := credentialtest.NewCredential(t, func(params *credential.RestoreParams) {
		params.UserID = usr.ID()
	})

	repo := helper.Repo()
	ctx := context.Background()
	require.NoError(t, repo.Add(ctx, cred))

	found, err := repo.FindByID(ctx, cred.ID())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, cred.ID(), found.ID())
}

func TestCredentialRepo_FindByUserAndKind_ReturnsNil_WhenNotExists(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	repo := helper.Repo()

	ctx := context.Background()
	cred, err := repo.FindByUserAndKind(ctx, user.NewID(), credential.KindPassword)

	assert.NoError(t, err)
	assert.Nil(t, cred)
}

func TestCredentialRepo_FindByUserAndKind_ReturnsCredential_WhenExists(t *testing.T) {
	helper := NewCredentialRepoTestHelper(t)
	usr := usertest.NewUser(t, nil)
	helper.SaveUser(usr)

	testCred := credentialtest.NewCredential(t, func(p *credential.RestoreParams) {
		p.UserID = usr.ID()
	})

	repo := helper.Repo()
	ctx := context.Background()
	require.NoError(t, repo.Add(ctx, testCred))

	found, err := repo.FindByUserAndKind(ctx, testCred.UserID(), testCred.Kind())

	assert.NoError(t, err)
	assert.NotNil(t, found)
	assert.Equal(t, testCred.ID(), found.ID())
}
