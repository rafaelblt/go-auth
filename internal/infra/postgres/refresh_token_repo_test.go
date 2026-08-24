package postgres_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type RefreshTokenRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
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

	usr := usertest.NewUser(helper.t, nil)
	postgrestest.InsertUser(helper.t, helper.db, usr)

	sess := sessiontest.NewSession(helper.t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	postgrestest.InsertSession(helper.t, helper.db, sess)

	token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.ExpiresAt = time.Now().UTC().Add(time.Hour)
	})
	postgrestest.InsertRefreshToken(helper.t, helper.db, token)

	return token
}

func (helper *RefreshTokenRepoTestHelper) CheckRefreshTokenExists(token *session.RefreshToken) bool {
	helper.t.Helper()
	return postgrestest.CheckRefreshTokenExists(helper.t, helper.db, token)
}

// TESTS

func TestNewRefreshTokenRepo_WithDBNil(t *testing.T) {
	uow, err := postgres.NewRefreshTokenRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, uow)
}

func TestNewRefreshTokenRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	uow, err := postgres.NewRefreshTokenRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
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

func TestRefreshTokenRepo_Update(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()
	require.NoError(t, token.Use(time.Now().UTC()))

	repo := helper.Repo()
	err := repo.Update(t.Context(), token)

	require.NoError(t, err)
	require.True(t, helper.CheckRefreshTokenExists(token))
}

func TestRefreshTokenRepo_Update_FailsWithTokenNonExistent(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := sessiontest.NewRefreshToken(t, nil)

	repo := helper.Repo()
	err := repo.Update(t.Context(), token)

	require.Error(t, err)
}

func TestRefreshTokenRepo_FindByHash_ReturnsNil_WhenIDNotExists(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	repo := helper.Repo()

	hash := sessiontest.MustRefreshTokenHash(helper.t, []byte{6, 7})
	usr, err := repo.FindByHash(t.Context(), hash)

	assert.NoError(t, err)
	assert.Nil(t, usr)
}

func TestRefreshTokenRepo_FindByHash_ReturnsToken_WhenIDExists(t *testing.T) {
	helper := NewRefreshTokenRepoTestHelper(t)
	token := helper.PersistentRefreshToken()

	repo := helper.Repo()
	found, err := repo.FindByHash(t.Context(), token.Hash())

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, token.ID(), found.ID())
	assert.True(t, token.Hash().Equal(found.Hash()), "token hash is different")
}
