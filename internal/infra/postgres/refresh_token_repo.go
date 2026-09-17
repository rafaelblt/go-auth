package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/domain/session"
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
			(id, session_id, parent_id, hash, expires_at, used_at, created_at, updated_at)
			VALUES
			(@id, @session_id, @parent_id, @hash, @expires_at, @used_at, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("refresh token insert failed: %w", err)
	}

	return nil
}

// MarkUsed guards the update with used_at IS NULL, so of two concurrent uses
// of the same token only one is applied. When nothing is updated, it checks
// whether the token exists, so a missing token is not reported as a reuse.
//
// See docs/development/decisions/0046-refresh-token-use-guarded-at-write.md.
func (repo *RefreshTokenRepo) MarkUsed(ctx context.Context, token *session.RefreshToken) error {
	model, err := mapRefreshTokenToModel(token)
	if err != nil {
		return fmt.Errorf("map refresh token to model failed: %w", err)
	}
	if model.UsedAt == nil {
		return errors.New("refresh token not used")
	}

	sql := `UPDATE refresh_tokens
			SET used_at = @used_at,
				updated_at = @updated_at
			WHERE id = @id AND used_at IS NULL`
	tag, err := repo.db.Exec(ctx, sql, pgx.StructArgs(model))

	if err != nil {
		return fmt.Errorf("refresh token mark used failed: %w", err)
	}
	if tag.RowsAffected() == 1 {
		return nil
	}

	exists, err := repo.existsByID(ctx, model.ID)
	if err != nil {
		return err
	}
	if !exists {
		return errors.New("refresh token not found")
	}

	return session.ErrTokenAlreadyUsed
}

func (repo *RefreshTokenRepo) existsByID(ctx context.Context, id string) (bool, error) {
	sql := "SELECT EXISTS(SELECT 1 FROM refresh_tokens WHERE id = $1)"

	var exists bool
	if err := repo.db.QueryRow(ctx, sql, id).Scan(&exists); err != nil {
		return false, fmt.Errorf("refresh token exists query failed: %w", err)
	}

	return exists, nil
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
