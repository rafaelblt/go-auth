package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
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

func (helper *SessionRepoTestHelper) SaveUser(usr *user.User) {
	helper.t.Helper()

	userRepo, err := postgres.NewUserRepo(helper.db)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userRepo.Add(context.Background(), usr))
}

func (helper *SessionRepoTestHelper) CheckSessionIsSaved(sess *session.Session) bool {
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
	helper.SaveUser(usr)

	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})

	repo := helper.Repo()
	err := repo.Add(context.Background(), sess)

	require.NoError(t, err)
	require.True(t, helper.CheckSessionIsSaved(sess))
}
