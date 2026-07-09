package e2e

import (
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type Asserts struct {
	pool *pgxpool.Pool
}

func (a *Asserts) RequireSessionIsRevoked(t *testing.T, id session.SessionID) {
	t.Helper()

	row := a.pool.QueryRow(t.Context(),
		`SELECT revoked_at FROM sessions WHERE id=$1`,
		id.Value(),
	)

	var revokedAt *time.Time
	err := row.Scan(&revokedAt)
	require.NotErrorIs(t, pgx.ErrNoRows, err, "session id not exists")
	require.NoError(t, err, "row scan failed")

	assert.NotNil(t, revokedAt, "session is not revoked")
}
