package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

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

func (repo *RefreshTokenRepo) Save(ctx context.Context, token *session.RefreshToken) error {
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
