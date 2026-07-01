package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/session"
)

type RefreshTokenRepo struct {
	db DB
}

func NewRefreshTokenRepo(db DB) (*RefreshTokenRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &RefreshTokenRepo{db: db}, nil
}

func (repo *RefreshTokenRepo) Add(ctx context.Context, token *session.RefreshToken) error {
	model, err := repo.mapToModel(token)
	if err != nil {
		return fmt.Errorf("map refresh token to model failed: %w", err)
	}

	sql := `INSERT INTO refresh_tokens
			(id, session_id, parent_id, hash, issued_at, expires_at, used_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)`
	_, err = repo.db.Exec(ctx, sql,
		model.ID,
		model.SessionID,
		model.ParentID,
		model.Hash,
		model.IssuedAt,
		model.ExpiresAt,
		model.UsedAt,
	)

	if err != nil {
		return fmt.Errorf("refresh token insert failed: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) Update(ctx context.Context, token *session.RefreshToken) error {
	model, err := repo.mapToModel(token)
	if err != nil {
		return fmt.Errorf("map refresh token to model failed: %w", err)
	}

	sql := `UPDATE refresh_tokens
			SET parent_id = $2
			SET used_at = $3
			WHERE id = $1`
	tag, err := repo.db.Exec(ctx, sql,
		model.ID,
		model.ParentID,
		model.UsedAt,
	)

	if err != nil {
		return fmt.Errorf("session update failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no rows affected in session update")
	}

	return nil
}

func (repo *RefreshTokenRepo) FindByHash(ctx context.Context, hash session.RefreshTokenHash) (*session.RefreshToken, error) {
	sql := "SELECT * FROM refresh_tokens WHERE hash = $1"

	rows, err := repo.db.Query(ctx, sql, hash.Value())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[refreshTokenModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	token, err := repo.mapToEntity(model)
	if err != nil {
		return nil, err
	}

	return token, nil
}

func (repo *RefreshTokenRepo) mapToModel(token *session.RefreshToken) (refreshTokenModel, error) {
	if token == nil {
		return refreshTokenModel{}, errors.New("refresh token nil")
	}
	if token.IsZero() {
		return refreshTokenModel{}, errors.New("refresh token zero")
	}

	id := token.ID().String()
	sessionID := token.SessionID().String()
	hash := token.Hash().Value()
	issuedAt := token.IssuedAt()
	expiresAt := token.ExpiresAt()

	var parentIDPtr *string
	var usedAtPtr *time.Time

	parentID, ok := token.ParentID()
	if ok {
		id := parentID.String()
		parentIDPtr = &id
	}

	usedAt, ok := token.UsedAt()
	if ok {
		usedAtPtr = &usedAt
	}

	model := refreshTokenModel{
		model:     model{ID: id},
		SessionID: sessionID,
		ParentID:  parentIDPtr,
		Hash:      hash,
		IssuedAt:  issuedAt,
		ExpiresAt: expiresAt,
		UsedAt:    usedAtPtr,
	}

	return model, nil
}

func (repo *RefreshTokenRepo) mapToEntity(model refreshTokenModel) (*session.RefreshToken, error) {
	id, err := session.ParseRefreshTokenID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse refresh token id failed: %w", err)
	}

	sessionID, err := session.ParseSessionID(model.SessionID)
	if err != nil {
		return nil, fmt.Errorf("parse session id failed: %w", err)
	}

	hash, err := session.NewRefreshTokenHash(model.Hash)
	if err != nil {
		return nil, fmt.Errorf("new refresh token hash failed: %w", err)
	}

	var parentID *session.RefreshTokenID
	if model.ParentID != nil {
		pID, err := session.ParseRefreshTokenID(*model.ParentID)
		if err != nil {
			return nil, fmt.Errorf("parse parent token id failed: %w", err)
		}
		parentID = &pID
	}

	token, err := session.RestoreRefreshToken(session.RefreshTokenRestoreParams{
		ID:        id,
		SessionID: sessionID,
		Hash:      hash,
		ParentID:  parentID,
		IssuedAt:  model.IssuedAt,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore refresh token failed: %w", err)
	}

	return token, nil
}
