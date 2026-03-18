package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

func Writer(t *testing.T) (infra.UserWriter, infra.PGDB) {
	ctx := context.Background()

	pool, err := infra.NewPool(ctx, testDB.ConnectionString())

	if err != nil {
		t.Fatal(err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}

	writer, err := infra.NewUserWriter(tx)
	return writer, tx
}

func TestSave(t *testing.T) {
	writer, db := Writer(t)
	ctx := context.Background()
	user := testutil.DefaultUser(t).Entity

	err := writer.Save(ctx, user)

	require.NoError(t, err)
	var exists bool
	err = db.QueryRow(ctx,
		`SELECT EXISTS(
			SELECT 1 FROM users WHERE
			id=$1 AND
			username=$2 AND
			status=$3 AND
			created_at=$4
			AND updated_at=$5
		)`,
		user.ID().Value(),
		user.Username().String(),
		user.Status().String(),
		user.CreatedAt(),
		user.UpdatedAt(),
	).Scan(&exists)
	require.True(t, exists)
}
