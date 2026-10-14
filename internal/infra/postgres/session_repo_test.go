package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPER

type SessionRepoTestHelper struct {
	t  *testing.T
	db *pgxpool.Pool
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
	return helper.PersistentSessionOf(helper.PersistentUser())
}

func (helper *SessionRepoTestHelper) PersistentSessionOf(usr *user.User) *session.Session {
	helper.t.Helper()
	return helper.PersistentSessionWith(usr, nil)
}

func (helper *SessionRepoTestHelper) PersistentSessionWith(
	usr *user.User, override func(p *session.SessionRestoreParams),
) *session.Session {
	helper.t.Helper()

	sess := sessiontest.NewSession(helper.t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
		if override != nil {
			override(p)
		}
	})
	postgrestest.InsertSession(helper.t, helper.db, sess)

	return sess
}

// PersistentChainOf stores a rotation chain of sess, one token per expiry, in
// order: each token is the parent of the next, and all but the last are used.
func (helper *SessionRepoTestHelper) PersistentChainOf(
	sess *session.Session, expiries ...time.Time,
) []*session.RefreshToken {
	helper.t.Helper()

	chain := make([]*session.RefreshToken, 0, len(expiries))
	for i, expiresAt := range expiries {
		token := sessiontest.NewRefreshToken(helper.t, func(p *session.RefreshTokenRestoreParams) {
			p.SessionID = sess.ID()
			p.Hash = sessiontest.NewRefreshTokenHash(helper.t, p.ID.String())
			p.ExpiresAt = expiresAt
			if i > 0 {
				p.ParentID = shared.Ptr(chain[i-1].ID())
			}
			if i < len(expiries)-1 {
				p.UsedAt = shared.Ptr(expiresAt.Add(-time.Minute))
			}
		})
		postgrestest.InsertRefreshToken(helper.t, helper.db, token)
		chain = append(chain, token)
	}

	return chain
}

func (helper *SessionRepoTestHelper) CheckSessionExists(sess *session.Session) bool {
	helper.t.Helper()
	return postgrestest.CheckSessionExists(helper.t, helper.db, sess)
}

func (helper *SessionRepoTestHelper) CheckRefreshTokenExists(token *session.RefreshToken) bool {
	helper.t.Helper()
	return postgrestest.CheckRefreshTokenExists(helper.t, helper.db, token)
}

// cutoffForTest is now at the precision of a timestamptz column, so a time
// equal to it is stored equal to it.
func cutoffForTest() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
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

func TestSessionRepo_RevokeAllByUserID_RevokesEveryActiveSessionOfTheUser(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	first := helper.PersistentSessionOf(usr)
	second := helper.PersistentSessionOf(usr)
	revokedAt := time.Now().UTC()

	repo := helper.Repo()
	err := repo.RevokeAllByUserID(t.Context(), usr.ID(), revokedAt)

	require.NoError(t, err)
	for _, sess := range []*session.Session{first, second} {
		sess.Revoke(revokedAt)
		assert.True(t, helper.CheckSessionExists(sess), "session not revoked at revokedAt")
	}
}

func TestSessionRepo_RevokeAllByUserID_KeepsTheTimeOfARevokedSession(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	revokedAt := time.Now().UTC()
	revoked := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
		p.RevokedAt = shared.Ptr(revokedAt.Add(-time.Hour))
		p.UpdatedAt = revokedAt.Add(-time.Hour)
	})
	postgrestest.InsertSession(t, helper.db, revoked)

	repo := helper.Repo()
	err := repo.RevokeAllByUserID(t.Context(), usr.ID(), revokedAt)

	require.NoError(t, err)
	assert.True(t, helper.CheckSessionExists(revoked), "earlier revocation was overwritten")
}

func TestSessionRepo_RevokeAllByUserID_LeavesOtherUsersSessionsActive(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	otherUser := usertest.NewUser(t, func(p *user.RestoreParams) {
		p.Username = usertest.MustUsername(t, "other_user")
	})
	postgrestest.InsertUser(t, helper.db, otherUser)
	other := helper.PersistentSessionOf(otherUser)

	repo := helper.Repo()
	err := repo.RevokeAllByUserID(t.Context(), usr.ID(), time.Now().UTC())

	require.NoError(t, err)
	assert.True(t, helper.CheckSessionExists(other), "another user's session changed")
}

func TestSessionRepo_DeleteDeadBefore_DeletesTheSessionsDeadBeforeTheCutoff(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	cutoff := cutoffForTest()
	before, after := cutoff.Add(-time.Hour), cutoff.Add(time.Hour)

	scenarios := []struct {
		desc      string
		createdAt time.Time
		revokedAt *time.Time
		expiries  []time.Time
		dead      bool
	}{
		{
			desc:      "revoked before the cutoff",
			revokedAt: &before,
			expiries:  []time.Time{after},
			dead:      true,
		},
		{
			desc:      "revoked at the cutoff",
			revokedAt: &cutoff,
			expiries:  []time.Time{after},
			dead:      false,
		},
		{
			desc:      "revoked after the cutoff",
			revokedAt: &after,
			expiries:  []time.Time{after},
			dead:      false,
		},
		{
			desc:     "latest token expired before the cutoff",
			expiries: []time.Time{before.Add(-time.Hour), before},
			dead:     true,
		},
		{
			desc:     "latest token expires at the cutoff",
			expiries: []time.Time{before, cutoff},
			dead:     false,
		},
		{
			desc:     "latest token expires after the cutoff",
			expiries: []time.Time{before, after},
			dead:     false,
		},
		{
			desc:      "no token, created before the cutoff",
			createdAt: before,
			dead:      true,
		},
		{
			desc:      "no token, created at the cutoff",
			createdAt: cutoff,
			dead:      false,
		},
	}

	sessions := make([]*session.Session, len(scenarios))
	chains := make([][]*session.RefreshToken, len(scenarios))
	dead := 0
	for i, sc := range scenarios {
		sessions[i] = helper.PersistentSessionWith(usr, func(p *session.SessionRestoreParams) {
			p.CreatedAt = cutoff.Add(-30 * 24 * time.Hour)
			if !sc.createdAt.IsZero() {
				p.CreatedAt = sc.createdAt
			}
			p.RevokedAt = sc.revokedAt
		})
		chains[i] = helper.PersistentChainOf(sessions[i], sc.expiries...)
		if sc.dead {
			dead++
		}
	}

	repo := helper.Repo()
	deleted, err := repo.DeleteDeadBefore(t.Context(), cutoff, 100)

	require.NoError(t, err)
	assert.Equal(t, dead, deleted)
	for i, sc := range scenarios {
		t.Run(sc.desc, func(t *testing.T) {
			assert.Equal(t, !sc.dead, helper.CheckSessionExists(sessions[i]), "session")
			for _, token := range chains[i] {
				assert.Equal(t, !sc.dead, helper.CheckRefreshTokenExists(token), "refresh token")
			}
		})
	}
}

// Reuse detection looks for the spent tokens of a live session, so none of them
// may go while the session lives.
func TestSessionRepo_DeleteDeadBefore_KeepsTheSpentAndExpiredTokensOfALiveSession(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	cutoff := cutoffForTest()
	sess := helper.PersistentSession()
	chain := helper.PersistentChainOf(sess,
		cutoff.Add(-48*time.Hour),
		cutoff.Add(-24*time.Hour),
		cutoff.Add(time.Hour),
	)

	repo := helper.Repo()
	deleted, err := repo.DeleteDeadBefore(t.Context(), cutoff, 100)

	require.NoError(t, err)
	assert.Zero(t, deleted)
	assert.True(t, helper.CheckSessionExists(sess), "live session deleted")
	for _, token := range chain {
		assert.True(t, helper.CheckRefreshTokenExists(token), "token of a live session deleted")
	}
}

func TestSessionRepo_DeleteDeadBefore_DeletesAtMostLimit(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	cutoff := cutoffForTest()
	revokedAt := cutoff.Add(-time.Hour)
	for range 3 {
		helper.PersistentSessionWith(usr, func(p *session.SessionRestoreParams) {
			p.RevokedAt = &revokedAt
		})
	}

	repo := helper.Repo()
	var batches []int
	for range 3 {
		deleted, err := repo.DeleteDeadBefore(t.Context(), cutoff, 2)
		require.NoError(t, err)
		batches = append(batches, deleted)
	}

	assert.Equal(t, []int{2, 1, 0}, batches)
}

func TestSessionRepo_DeleteDeadBefore_SkipsASessionAnotherTransactionHolds(t *testing.T) {
	helper := NewSessionRepoTestHelper(t)
	usr := helper.PersistentUser()
	cutoff := cutoffForTest()
	revokedAt := cutoff.Add(-time.Hour)
	held := helper.PersistentSessionWith(usr, func(p *session.SessionRestoreParams) {
		p.RevokedAt = &revokedAt
	})
	free := helper.PersistentSessionWith(usr, func(p *session.SessionRestoreParams) {
		p.RevokedAt = &revokedAt
	})

	tx, err := helper.db.Begin(t.Context())
	require.NoError(t, err)
	// t.Context() is already canceled when cleanups run.
	t.Cleanup(func() { tx.Rollback(context.Background()) })
	_, err = tx.Exec(t.Context(), "SELECT 1 FROM sessions WHERE id = $1 FOR UPDATE", held.ID().Value())
	require.NoError(t, err)

	repo := helper.Repo()
	ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
	defer cancel()
	deleted, err := repo.DeleteDeadBefore(ctx, cutoff, 100)

	require.NoError(t, err, "the delete waited on the held session")
	assert.Equal(t, 1, deleted)
	assert.False(t, helper.CheckSessionExists(free), "free session kept")
	assert.True(t, helper.CheckSessionExists(held), "held session deleted")

	require.NoError(t, tx.Rollback(t.Context()))
	deleted, err = repo.DeleteDeadBefore(t.Context(), cutoff, 100)

	require.NoError(t, err)
	assert.Equal(t, 1, deleted, "held session not deleted once released")
}

func TestSessionRepo_DeleteDeadBefore_Fails_WithZeroCutoffOrLimit(t *testing.T) {
	testCases := []struct {
		desc   string
		cutoff time.Time
		limit  int
	}{
		{desc: "zero cutoff", cutoff: time.Time{}, limit: 100},
		{desc: "zero limit", cutoff: time.Now().UTC(), limit: 0},
		{desc: "negative limit", cutoff: time.Now().UTC(), limit: -1},
	}
	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			helper := NewSessionRepoTestHelper(t)

			repo := helper.Repo()
			deleted, err := repo.DeleteDeadBefore(t.Context(), tC.cutoff, tC.limit)

			assert.Error(t, err)
			assert.Zero(t, deleted)
		})
	}
}
