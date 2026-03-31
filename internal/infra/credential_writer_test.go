package infra_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

// HELPER

type CredentialWriterTestHelper struct {
	t    *testing.T
	pool *pgxpool.Pool
}
func NewCredentialWriterTestHelper(t *testing.T) CredentialWriterTestHelper {
	ctx := context.Background()
	pool, err := infra.NewPool(ctx, testDB.ConnectionString())
	require.NoError(t, err)
	return CredentialWriterTestHelper{t, pool}
}
func (helper CredentialWriterTestHelper) WriterAndDB() (infra.CredentialWriter, infra.PGDB) {
	writer, err := infra.NewCredentialWriter(helper.pool)
	require.NoError(helper.t, err)
	return writer, helper.pool
}
func (helper CredentialWriterTestHelper) PersistentUser() *domain.User {
	user := testutil.DefaultUser(helper.t)

	userWriter, err := infra.NewUserWriter(helper.pool)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userWriter.Save(context.Background(), user))

	return user
}

// TESTS

func TestCredentialWriter_Save(t *testing.T) {
	helper := NewCredentialWriterTestHelper(t)
	writer, db := helper.WriterAndDB()
	persistentUser := helper.PersistentUser()
	credential := testutil.PasswordCredential(t,
		testutil.WithUserID(persistentUser.ID()),
	)

	err := writer.Save(context.Background(), credential)

	require.NoError(t, err)
	var exists bool
	err = db.QueryRow(context.Background(),
		`SELECT EXISTS(
			SELECT 1 FROM credentials WHERE
			id=$1 AND
			user_id=$2 AND
			kind=$3 AND
			provider=$4 AND
			secret=$5 AND
			created_at=$6 AND
			updated_at=$7
		)`,
		credential.ID().Value(),
		credential.UserID().Value(),
		credential.Kind().String(),
		credential.Provider().String(),
		credential.Secret().Value(),
		credential.CreatedAt(),
		credential.UpdatedAt(),
	).Scan(&exists)
	require.True(t, exists)
}
