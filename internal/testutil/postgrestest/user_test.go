package postgrestest_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rafaelblt/go-auth/internal/infra/postgres"
	"github.com/rafaelblt/go-auth/internal/testutil/postgrestest"
	"github.com/rafaelblt/go-auth/internal/testutil/usertest"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type userRow struct {
	Username  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func selectUserRow(t *testing.T, db postgres.DB, id uuid.UUID) userRow {
	t.Helper()

	var row userRow
	err := db.QueryRow(t.Context(),
		`SELECT username, status, created_at, updated_at FROM users WHERE id=$1`,
		id,
	).Scan(&row.Username, &row.Status, &row.CreatedAt, &row.UpdatedAt)
	require.NoError(t, err, "user select failed")

	row.CreatedAt = row.CreatedAt.UTC()
	row.UpdatedAt = row.UpdatedAt.UTC()
	return row
}

func TestInsertUser(t *testing.T) {
	db := poolFactory.AcquireWithMigrations(t)
	usr := usertest.NewUser(t, nil)

	postgrestest.InsertUser(t, db, usr)

	expected := userRow{
		Username:  usr.Username().String(),
		Status:    usr.Status().String(),
		CreatedAt: usr.CreatedAt(),
		UpdatedAt: usr.UpdatedAt(),
	}
	assert.Equal(t, expected, selectUserRow(t, db, usr.ID().Value()))
}
