package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

type UserWriterTestHelper struct {
	t *testing.T
}

func NewUserWriterTestHelper(t *testing.T) UserWriterTestHelper {
	return UserWriterTestHelper{t}
}

func (helper UserWriterTestHelper) DB() infra.PGDB {
	return testDB.NewTx(helper.t)
}

func (helper UserWriterTestHelper) Writer(db infra.PGDB) infra.UserWriter {
	helper.t.Helper()
	writer, err := infra.NewUserWriter(db)
	require.NoError(helper.t, err)
	return writer
}

func TestUserWriter_Save(t *testing.T) {
	helper := NewUserWriterTestHelper(t)
	db := helper.DB()
	writer := helper.Writer(db)
	user := testutil.DefaultUser(t)

	err := writer.Save(context.Background(), user)

	require.NoError(t, err)
	var exists bool
	err = db.QueryRow(context.Background(),
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
