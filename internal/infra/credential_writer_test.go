package infra_test

import (
	"context"
	"testing"

	"github.com/rafaelblt/go-auth/internal/domain"
	"github.com/rafaelblt/go-auth/internal/infra"
	"github.com/rafaelblt/go-auth/internal/testutil"
	"github.com/stretchr/testify/require"
)

// HELPER

type CredentialWriterTestHelper struct {
	t  *testing.T
}

func NewCredentialWriterTestHelper(t *testing.T) CredentialWriterTestHelper {
	return CredentialWriterTestHelper{t}
}

func (helper CredentialWriterTestHelper) DB() infra.PGDB {
	return testDB.NewTx(helper.t)
}

func (helper CredentialWriterTestHelper) Writer(db infra.PGDB) infra.CredentialWriter {
	helper.t.Helper()

	writer, err := infra.NewCredentialWriter(db)
	require.NoError(helper.t, err)

	return writer
}

func (helper CredentialWriterTestHelper) PersistentUser(db infra.PGDB) *domain.User {
	helper.t.Helper()
	require.NotNil(helper.t, db)

	user := testutil.DefaultUser(helper.t)

	userWriter, err := infra.NewUserWriter(db)
	require.NoError(helper.t, err)

	require.NoError(helper.t, userWriter.Save(context.Background(), user))

	return user
}

// TESTS

func TestCredentialWriter_Save(t *testing.T) {
	helper := NewCredentialWriterTestHelper(t)
	db := helper.DB()
	writer := helper.Writer(db)
	persistentUser := helper.PersistentUser(db)

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
