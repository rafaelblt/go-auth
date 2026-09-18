package postgrestest_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/sessiontest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// HELPERS

type refreshTokenRow struct {
	SessionID uuid.UUID
	ParentID  *uuid.UUID
	Hash      []byte
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func selectRefreshTokenRow(t *testing.T, db postgres.DB, id uuid.UUID) refreshTokenRow {
	t.Helper()

	var row refreshTokenRow
	err := db.QueryRow(t.Context(),
		`SELECT session_id, parent_id, hash, expires_at, used_at, created_at, updated_at
		FROM refresh_tokens WHERE id=$1`,
		id,
	).Scan(&row.SessionID, &row.ParentID, &row.Hash, &row.ExpiresAt, &row.UsedAt, &row.CreatedAt, &row.UpdatedAt)
	require.NoError(t, err, "refresh token select failed")

	row.ExpiresAt = row.ExpiresAt.UTC()
	row.UsedAt = utcPtr(row.UsedAt)
	row.CreatedAt = row.CreatedAt.UTC()
	row.UpdatedAt = row.UpdatedAt.UTC()
	return row
}

func expectedRefreshTokenRow(token *session.RefreshToken) refreshTokenRow {
	var parentID *uuid.UUID
	if pID, ok := token.ParentID(); ok {
		parentID = shared.Ptr(pID.Value())
	}

	return refreshTokenRow{
		SessionID: token.SessionID().Value(),
		ParentID:  parentID,
		Hash:      token.Hash().Value(),
		ExpiresAt: token.ExpiresAt(),
		UsedAt:    shared.PtrFromOk(token.UsedAt()),
		CreatedAt: token.CreatedAt(),
		UpdatedAt: token.UpdatedAt(),
	}
}

func persistentSession(t *testing.T, db postgres.DB) *session.Session {
	t.Helper()
	usr := persistentUser(t, db)
	sess := sessiontest.NewSession(t, func(p *session.SessionRestoreParams) {
		p.UserID = usr.ID()
	})
	postgrestest.InsertSession(t, db, sess)
	return sess
}

// refreshTokenPair builds a used root token and its unused child, both bound to sess.
func refreshTokenPair(t *testing.T, sess *session.Session) (root, child *session.RefreshToken) {
	t.Helper()

	root = sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.Hash = sessiontest.NewRefreshTokenHash(t, "root")
		p.UsedAt = shared.Ptr(time.Date(1990, 4, 24, 12, 33, 0, 0, time.UTC))
	})
	child = sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		p.SessionID = sess.ID()
		p.Hash = sessiontest.NewRefreshTokenHash(t, "child")
		p.ParentID = shared.Ptr(root.ID())
	})

	return root, child
}

// refreshTokenWith copies token and applies override on the copy.
func refreshTokenWith(t *testing.T, token *session.RefreshToken, override func(p *session.RefreshTokenRestoreParams)) *session.RefreshToken {
	t.Helper()
	return sessiontest.NewRefreshToken(t, func(p *session.RefreshTokenRestoreParams) {
		*p = session.RefreshTokenRestoreParams{
			ID:        token.ID(),
			SessionID: token.SessionID(),
			Hash:      token.Hash(),
			ParentID:  shared.PtrFromOk(token.ParentID()),
			ExpiresAt: token.ExpiresAt(),
			UsedAt:    shared.PtrFromOk(token.UsedAt()),
			CreatedAt: token.CreatedAt(),
			UpdatedAt: token.UpdatedAt(),
		}
		override(p)
	})
}

// TESTS

func TestInsertRefreshToken(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	root, child := refreshTokenPair(t, persistentSession(t, db))

	postgrestest.InsertRefreshToken(t, db, root)
	postgrestest.InsertRefreshToken(t, db, child)

	assert.Equal(t, expectedRefreshTokenRow(root), selectRefreshTokenRow(t, db, root.ID().Value()),
		"root token: without parent and used")
	assert.Equal(t, expectedRefreshTokenRow(child), selectRefreshTokenRow(t, db, child.ID().Value()),
		"child token: with parent and not used")
}

func TestCheckRefreshTokenExists(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	sess := persistentSession(t, db)
	root, child := refreshTokenPair(t, sess)
	postgrestest.InsertRefreshToken(t, db, root)
	postgrestest.InsertRefreshToken(t, db, child)

	tests := []struct {
		name     string
		token    *session.RefreshToken
		expected bool
	}{
		{"matches root token", root, true},
		{"matches child token", child, true},
		{"id differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.ID = session.NewRefreshTokenID()
		}), false},
		{"session id differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.SessionID = session.NewSessionID()
		}), false},
		{"parent id set but stored null", refreshTokenWith(t, root, func(p *session.RefreshTokenRestoreParams) {
			p.ParentID = shared.Ptr(child.ID())
		}), false},
		{"parent id null but stored set", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.ParentID = nil
		}), false},
		{"parent id differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.ParentID = shared.Ptr(session.NewRefreshTokenID())
		}), false},
		{"hash differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.Hash = sessiontest.NewRefreshTokenHash(t, "other")
		}), false},
		{"expires at differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.ExpiresAt = p.ExpiresAt.Add(time.Second)
		}), false},
		{"used at set but stored null", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.UsedAt = shared.Ptr(p.CreatedAt)
		}), false},
		{"used at null but stored set", refreshTokenWith(t, root, func(p *session.RefreshTokenRestoreParams) {
			p.UsedAt = nil
		}), false},
		{"used at differs", refreshTokenWith(t, root, func(p *session.RefreshTokenRestoreParams) {
			p.UsedAt = shared.Ptr(p.UsedAt.Add(time.Second))
		}), false},
		{"created at differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.CreatedAt = p.CreatedAt.Add(time.Second)
		}), false},
		{"updated at differs", refreshTokenWith(t, child, func(p *session.RefreshTokenRestoreParams) {
			p.UpdatedAt = p.UpdatedAt.Add(time.Second)
		}), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := postgrestest.CheckRefreshTokenExists(t, db, tt.token)
			assert.Equal(t, tt.expected, result)
		})
	}
}
