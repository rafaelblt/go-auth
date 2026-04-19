package testutil

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/require"
)

type IsolatedDB struct {
	db *Database
}

func NewIsolatedDB(ctx context.Context) (*IsolatedDB, error) {
	db, err := NewDatabase(ctx)
	if err != nil {
		return nil, err
	}

	return &IsolatedDB{db}, nil
}

func (idb *IsolatedDB) TxForTest(t *testing.T) pgx.Tx {
	t.Helper()

	tx, err := idb.db.Pool().Begin(context.Background())
	require.NoError(t, err)

	nestedTx, err := tx.Begin(context.Background()) 
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = tx.Rollback(context.Background())
	})

	return nestedTx
}

func (db *IsolatedDB) Close(ctx context.Context) error {
	return db.Close(ctx)
}