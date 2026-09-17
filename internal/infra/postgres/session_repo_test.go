package postgres_test

import (
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type SessionRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
}

func NewSessionRepoTestHelper(t *testing.T) *SessionRepoTestHelper {
	db := poolFactory.AcquireWithMigrations(t)
	return &SessionRepoTestHelper{t, db}
}

func (helper *SessionRepoTestHelper) Repo() *postgres.SessionRepo {
	helper.t.Helper()

	repo, err := postgres.NewSessionRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper *SessionRepoTestHelper) PersistentUser() *user.User {
	helper.t.Helper()
	usr := usertest.NewUser(helper.t, nil)
	postgrestest.InsertUser(helper.t, helper.db, usr)
	return usr
}

func (helper *SessionRepoTestHelper) PersistentSession() *session.Session {
	helper.t.Helper()

	usr := usertest.NewUser(helper.t, nil)
	postgrestest.InsertUser(helper.t, helper.db, usr)

	sess := sessiontest.NewSession(helper.t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	postgrestest.InsertSession(helper.t, helper.db, sess)

	return sess
}

func (helper *SessionRepoTestHelper) CheckSessionExists(sess *session.Session) bool {
	helper.t.Helper()
	return postgrestest.CheckSessionExists(helper.t, helper.db, sess)
}

// TESTS

func TestNewSessionRepo_WithDBNil(t *testing.T) {
	uow, err := postgres.NewSessionRepo(nil)
	assert.Error(t, err)
	assert.Zero(t, uow)
}

func TestNewSessionRepo_WithValidDB(t *testing.T) {
	db := poolFactory.Acquire(t)

	uow, err := postgres.NewSessionRepo(db)

	assert.NoError(t, err)
	assert.NotNil(t, uow)
}

func TestSessionRepo_Add(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})

	repo := helper.Repo()
	err := repo.Add(t.Context(), sess)

	require.NoError(t, err)
	require.True(t, helper.CheckSessionExists(sess))
}

func TestSessionRepo_Update(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	sess := helper.PersistentSession()
	sess.Revoke(time.Now().UTC())

	repo := helper.Repo()
	err := repo.Update(t.Context(), sess)

	require.NoError(t, err)
	require.True(t, helper.CheckSessionExists(sess))
}

func TestSessionRepo_Update_FailsWithSessionNonExistent(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	sess := sessiontest.NewSession(t, nil)

	repo := helper.Repo()
	err := repo.Update(t.Context(), sess)

	require.Error(t, err)
}

func TestSessionRepo_FindByID_ReturnsNil_WhenIDNotExists(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	repo := helper.Repo()

	usr, err := repo.FindByID(t.Context(), session.NewSessionID())

	assert.NoError(t, err)
	assert.Nil(t, usr)
}

func TestSessionRepo_FindByID_ReturnsSession_WhenIDExists(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	sess := helper.PersistentSession()

	repo := helper.Repo()
	found, err := repo.FindByID(t.Context(), sess.ID())

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, sess.ID(), found.ID())
}
