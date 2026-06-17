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

func (repo CredentialRepo) Save(ctx context.Context, cred *credential.Credential) error {
	if cred == nil {
		return errors.New("cannot save a nil credential in database")
	}
	if cred.IsZero() {
		return errors.New("cannot save a zero credential in database")
	}

	sql := `INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := repo.db.Exec(ctx, sql,
		cred.ID().Value(),
		cred.UserID().Value(),
		cred.Kind().String(),
		cred.Provider().String(),
		cred.Secret().Value(),
		cred.CreatedAt(),
		cred.UpdatedAt(),
	)

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

	cred, err := repo.mapModel(model)
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

	cred, err := repo.mapModel(model)
	if err != nil {
		return nil, err
	}

	return cred, nil
}

func (repo *CredentialRepo) mapModel(model credentialModel) (*credential.Credential, error) {
	id, err := credential.ParseID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("credential id parse failed: %w", err)
	}
	userID, err := user.ParseID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("user id parse failed: %w", err)
	}
	kind, err := credential.ParseKind(model.Kind)
	if err != nil {
		return nil, fmt.Errorf("kind validation failed: %w", err)
	}
	provider, err := credential.ParseProvider(model.Provider)
	if err != nil {
		return nil, fmt.Errorf("provider creation failed: %w", err)
	}
	secret, err := credential.NewSecret(model.Secret)
	if err != nil {
		return nil, fmt.Errorf("secret validation failed: %w", err)
	}

	usr, err := credential.RestoreCredential(credential.RestoreParams{
		ID:        id,
		UserID:    userID,
		Kind:      kind,
		Provider:  provider,
		Secret:    secret,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore credential failed: %w", err)
	}

	return usr, nil
}
