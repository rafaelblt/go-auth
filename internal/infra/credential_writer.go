package infra

import (
	"context"
	"errors"

	"github.com/rafaelblt/go-auth/internal/domain"
)

type CredentialWriter struct {
	db PGDB
}

func NewCredentialWriter(db PGDB) (CredentialWriter, error) {
	if db == nil {
		return CredentialWriter{}, errors.New("PGDB cannot be nil")
	}
	return CredentialWriter{db: db}, nil
}

func (writer CredentialWriter) Save(ctx context.Context, cred *domain.Credential) error {
	model, err := MapCredentialDomainToModel(cred)
	if err != nil {
		return err
	}
	_, err = writer.db.Exec(ctx,
		`INSERT INTO credentials
			(id, user_id, kind, provider, secret, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		model.ID,
		model.UserID,
		model.Kind,
		model.Provider,
		model.Secret,
		model.CreatedAt,
		model.UpdatedAt,
	)
	if err != nil {
		return err
	}
	return nil
}
