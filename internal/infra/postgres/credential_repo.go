package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/user"
)

type CredentialRepo struct {
	db DB
}

func NewCredentialRepo(db DB) (*CredentialRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &CredentialRepo{db: db}, nil
}

func (repo *CredentialRepo) Add(ctx context.Context, cred *credential.Credential) error {
	model, err := mapCredentialToModel(cred)
	if err != nil {
		return fmt.Errorf("map credential to model failed: %w", err)
	}

	sql := `INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
			VALUES
			(@id, @user_id, @kind, @provider, @secret, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("credential insert failed: %w", err)
	}

	return nil
}

func (repo *CredentialRepo) FindByID(ctx context.Context, id credential.ID) (*credential.Credential, error) {
	sql := "SELECT * FROM credentials WHERE id = $1"

	rows, err := repo.db.Query(ctx, sql, id.Value().String())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[credentialModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	cred, err := mapCredentialToEntity(model)
	if err != nil {
		return nil, err
	}

	return cred, nil
}

func (repo *CredentialRepo) FindByUserAndKind(
	ctx context.Context, userID user.ID, kind credential.Kind,
) (*credential.Credential, error) {
	sql := `SELECT * FROM credentials
			WHERE user_id = $1 AND kind = $2`

	rows, err := repo.db.Query(ctx, sql, userID.Value().String(), kind.String())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[credentialModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	cred, err := mapCredentialToEntity(model)
	if err != nil {
		return nil, err
	}

	return cred, nil
}
