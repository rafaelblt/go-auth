package postgres

import (
	"context"
	"errors"
	"fmt"

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
	model, err := mapRefreshTokenToModel(token)
	if err != nil {
		return fmt.Errorf("map refresh token to model failed: %w", err)
	}

	sql := `INSERT INTO refresh_tokens
			(id, session_id, parent_id, hash, created_at, expires_at, used_at)
			VALUES
			(@id, @session_id, @parent_id, @hash, @created_at, @expires_at, @used_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("refresh token insert failed: %w", err)
	}

	return nil
}

func (repo *RefreshTokenRepo) Update(ctx context.Context, token *session.RefreshToken) error {
	model, err := mapRefreshTokenToModel(token)
	if err != nil {
		return fmt.Errorf("map refresh token to model failed: %w", err)
	}

	sql := `UPDATE refresh_tokens
			SET parent_id = @parent_id,
				used_at = @used_at
			WHERE id = @id`
	tag, err := repo.db.Exec(ctx, sql, pgx.StructArgs(model))

	if err != nil {
		return fmt.Errorf("refresh token update failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no rows affected in refresh token update")
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

	token, err := mapRefreshTokenToEntity(model)
	if err != nil {
		return nil, err
	}

	return token, nil
}
