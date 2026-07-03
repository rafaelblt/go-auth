package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type SessionRepoTestHelper struct {
	t  *testing.T
	db postgres.DB
}

func NewSessionRepoTestHelper(t *testing.T) *SessionRepoTestHelper {
	db := poolFactory.Acquire(t)
	return &SessionRepoTestHelper{t, db}
}

func (helper *SessionRepoTestHelper) Repo() *postgres.SessionRepo {
	helper.t.Helper()

	repo, err := postgres.NewSessionRepo(helper.db)
	require.NoError(helper.t, err)

	return repo
}

func (helper *SessionRepoTestHelper) AddUser(usr *user.User) {
	helper.t.Helper()
	postgrestest.InsertUser(helper.t, helper.db, usr)
}

func (helper *SessionRepoTestHelper) GetSession() *session.Session {
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
	var result bool

	var revokedAtPtr *time.Time
	revokedAt, ok := sess.RevokedAt()
	if ok {
		revokedAtPtr = &revokedAt
	}

	err := helper.db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM sessions WHERE
			id=$1 AND
			user_id=$2 AND
			issued_at=$3 AND
			revoked_at IS NOT DISTINCT FROM $4
		)`,
		sess.ID().Value(),
		sess.UserID().Value(),
		sess.IssuedAt(),
		revokedAtPtr,
	).Scan(&result)
	require.NoError(helper.t, err)
	return result
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

func TestSessionRepo_Save(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)

	usr := usertest.NewUser(t, nil)
	helper.AddUser(usr)
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

	sess := helper.GetSession()
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
	sess := helper.GetSession()

	repo := helper.Repo()
	found, err := repo.FindByID(t.Context(), sess.ID())

	require.NoError(t, err)
	require.NotNil(t, found)
	assert.Equal(t, found.ID(), sess.ID())
	assert.Equal(t, found.UserID(), sess.UserID())
	assert.Equal(t, found.IssuedAt(), sess.IssuedAt())
	actualRevokedAt, _ := found.RevokedAt()
	expectedRevokedAt, _ := sess.RevokedAt()
	assert.Equal(t, actualRevokedAt, expectedRevokedAt)
}
