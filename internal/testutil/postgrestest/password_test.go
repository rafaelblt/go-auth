package postgrestest_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/testutil/passwordtest"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type passwordRow struct {
	UserID    uuid.UUID
	Hash      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func selectPasswordRow(t *testing.T, db postgres.DB, id uuid.UUID) passwordRow {
	t.Helper()

	var row passwordRow
	err := db.QueryRow(t.Context(),
		`SELECT user_id, hash, created_at, updated_at FROM passwords WHERE id=$1`,
		id,
	).Scan(&row.UserID, &row.Hash, &row.CreatedAt, &row.UpdatedAt)
	require.NoError(t, err, "password select failed")

	row.CreatedAt = row.CreatedAt.UTC()
	row.UpdatedAt = row.UpdatedAt.UTC()
	return row
}

func TestInsertPassword(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	usr := usertest.NewUser(t, nil)
	postgrestest.InsertUser(t, db, usr)
	pwd := passwordtest.NewPassword(t, func(p *password.RestoreParams) {
		p.UserID = usr.ID()
	})

	postgrestest.InsertPassword(t, db, pwd)

	expected := passwordRow{
		UserID:    pwd.UserID().Value(),
		Hash:      pwd.Hash().Value(),
		CreatedAt: pwd.CreatedAt(),
		UpdatedAt: pwd.UpdatedAt(),
	}
	assert.Equal(t, expected, selectPasswordRow(t, db, pwd.ID().Value()))
}
