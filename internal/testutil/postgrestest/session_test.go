package postgrestest_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/rafaelblt/go-auth/internal/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

type sessionRow struct {
	UserID    uuid.UUID
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func selectSessionRow(t *testing.T, db postgres.DB, id uuid.UUID) sessionRow {
	t.Helper()

	var row sessionRow
	err := db.QueryRow(t.Context(),
		`SELECT user_id, revoked_at, created_at, updated_at FROM sessions WHERE id=$1`,
		id,
	).Scan(&row.UserID, &row.RevokedAt, &row.CreatedAt, &row.UpdatedAt)
	require.NoError(t, err, "session select failed")

	row.RevokedAt = utcPtr(row.RevokedAt)
	row.CreatedAt = row.CreatedAt.UTC()
	row.UpdatedAt = row.UpdatedAt.UTC()
	return row
}

func persistentUser(t *testing.T, db postgres.DB) *user.User {
	t.Helper()
	usr := usertest.NewUser(t, nil)
	postgrestest.InsertUser(t, db, usr)
	return usr
}

// sessionWith copies sess and applies override on the copy.
func sessionWith(t *testing.T, sess *session.Session, override func(p *session.SessionRestoreParams)) *session.Session {
	t.Helper()
	return sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		*p = session.SessionRestoreParams{
			ID:        sess.ID(),
			UserID:    sess.UserID(),
			RevokedAt: shared.PtrFromOk(sess.RevokedAt()),
			CreatedAt: sess.CreatedAt(),
			UpdatedAt: sess.UpdatedAt(),
		}
		override(p)
	})
}

// TESTS

func TestInsertSession(t *testing.T) {
	tests := []struct {
		name      string
		revokedAt *time.Time
	}{
		{"not revoked", nil},
		{"revoked", shared.Ptr(time.Date(1970, 4, 13, 3, 7, 0, 0, time.UTC))},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			db := poolFactory.AcquireWithMigrations(t)
			usr := persistentUser(t, db)
			sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
				p.UserID = usr.ID()
				p.RevokedAt = tt.revokedAt
			})

			postgrestest.InsertSession(t, db, sess)

			expected := sessionRow{
				UserID:    sess.UserID().Value(),
				RevokedAt: shared.PtrFromOk(sess.RevokedAt()),
				CreatedAt: sess.CreatedAt(),
				UpdatedAt: sess.UpdatedAt(),
			}
			assert.Equal(t, expected, selectSessionRow(t, db, sess.ID().Value()))
		})
	}
}

func TestCheckSessionExists(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	usr := persistentUser(t, db)

	active := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	revoked := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
		p.RevokedAt = shared.Ptr(time.Date(1970, 4, 13, 3, 7, 0, 0, time.UTC))
	})
	postgrestest.InsertSession(t, db, active)
	postgrestest.InsertSession(t, db, revoked)

	tests := []struct {
		name     string
		sess     *session.Session
		expected bool
	}{
		{"matches active session", active, true},
		{"matches revoked session", revoked, true},
		{"id differs", sessionWith(t, active, func(p *session.SessionRestoreParams) {
			p.ID = session.NewSessionID()
		}), false},
		{"user id differs", sessionWith(t, active, func(p *session.SessionRestoreParams) {
			p.UserID = user.NewID()
		}), false},
		{"revoked at set but stored null", sessionWith(t, active, func(p *session.SessionRestoreParams) {
			p.RevokedAt = shared.Ptr(p.CreatedAt)
		}), false},
		{"revoked at null but stored set", sessionWith(t, revoked, func(p *session.SessionRestoreParams) {
			p.RevokedAt = nil
		}), false},
		{"revoked at differs", sessionWith(t, revoked, func(p *session.SessionRestoreParams) {
			p.RevokedAt = shared.Ptr(p.RevokedAt.Add(time.Second))
		}), false},
		{"created at differs", sessionWith(t, active, func(p *session.SessionRestoreParams) {
			p.CreatedAt = p.CreatedAt.Add(time.Second)
		}), false},
		{"updated at differs", sessionWith(t, active, func(p *session.SessionRestoreParams) {
			p.UpdatedAt = p.UpdatedAt.Add(time.Second)
		}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postgrestest.CheckSessionExists(t, db, tt.sess)
			assert.Equal(t, tt.expected, result)
		})
	}
}
