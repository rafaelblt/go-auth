package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/port"
)

type SigningKeyRepo struct {
	db DB
}

func NewSigningKeyRepo(db DB) (*SigningKeyRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &SigningKeyRepo{db: db}, nil
}

func (repo *SigningKeyRepo) List(ctx context.Context) ([]port.StoredSigningKey, error) {
	sql := "SELECT generation, seed, active_at, created_at FROM signing_keys ORDER BY generation"

	rows, err := repo.db.Query(ctx, sql)
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	models, err := pgx.CollectRows(rows, pgx.RowToStructByName[signingKeyModel])
	if err != nil {
		return nil, fmt.Errorf("collect rows failed: %w", err)
	}

	keys := make([]port.StoredSigningKey, 0, len(models))
	for _, model := range models {
		keys = append(keys, mapSigningKeyToRecord(model))
	}

	return keys, nil
}

// Add absorbs a conflict on the generation on purpose: the caller lists the
// keys again either way, so it needs what is stored, not who stored it.
func (repo *SigningKeyRepo) Add(ctx context.Context, key port.StoredSigningKey) error {
	model := mapSigningKeyToModel(key)

	sql := `INSERT INTO signing_keys
			(generation, seed, active_at, created_at)
			VALUES
			(@generation, @seed, @active_at, @created_at)
			ON CONFLICT (generation) DO NOTHING`
	_, err := repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("signing key insert failed: %w", err)
	}

	return nil
}

func (repo *SigningKeyRepo) DeleteBefore(ctx context.Context, generation int64) error {
	sql := "DELETE FROM signing_keys WHERE generation < $1"

	if _, err := repo.db.Exec(ctx, sql, generation); err != nil {
		return fmt.Errorf("signing key delete failed: %w", err)
	}

	return nil
}
