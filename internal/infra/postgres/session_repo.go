package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type SessionRepo struct {
	db DB
}

func NewSessionRepo(db DB) (*SessionRepo, error) {
	if db == nil {
		return nil, errors.New("db nil")
	}
	return &SessionRepo{db: db}, nil
}

func (repo *SessionRepo) Add(ctx context.Context, sess *session.Session) error {
	model, err := mapSessionToModel(sess)
	if err != nil {
		return fmt.Errorf("map session to model failed: %w", err)
	}

	sql := `INSERT INTO sessions
			(id, user_id, revoked_at, created_at, updated_at)
			VALUES
			(@id, @user_id, @revoked_at, @created_at, @updated_at)`
	_, err = repo.db.Exec(ctx, sql, pgx.StrictStructArgs(model))

	if err != nil {
		return fmt.Errorf("session insert failed: %w", err)
	}

	return nil
}

func (repo *SessionRepo) Update(ctx context.Context, sess *session.Session) error {
	model, err := mapSessionToModel(sess)
	if err != nil {
		return fmt.Errorf("map session to model failed: %w", err)
	}

	sql := `UPDATE sessions
			SET revoked_at = @revoked_at,
				updated_at = @updated_at
			WHERE id = @id`
	tag, err := repo.db.Exec(ctx, sql, pgx.StructArgs(model))

	if err != nil {
		return fmt.Errorf("session update failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no rows affected in session update")
	}

	return nil
}

// RevokeAllByUserID revokes in SQL rather than through Session.Revoke, so the
// sessions the caller never read are revoked too. revoked_at IS NULL keeps the
// time of an earlier revocation, as Session.Revoke does in memory.
func (repo *SessionRepo) RevokeAllByUserID(
	ctx context.Context, userID user.ID, revokedAt time.Time,
) error {
	if userID.IsZero() {
		return errors.New("user id zero")
	}

	sql := `UPDATE sessions
			SET revoked_at = $2,
				updated_at = $2
			WHERE user_id = $1 AND revoked_at IS NULL`
	_, err := repo.db.Exec(ctx, sql, userID.Value(), revokedAt)

	if err != nil {
		return fmt.Errorf("sessions revoke by user id failed: %w", err)
	}

	return nil
}

func (repo *SessionRepo) FindByID(ctx context.Context, id session.SessionID) (*session.Session, error) {
	sql := "SELECT * FROM sessions WHERE id = $1"

	rows, err := repo.db.Query(ctx, sql, id.Value())
	if err != nil {
		return nil, fmt.Errorf("db query failed: %w", err)
	}
	defer rows.Close()

	model, err := pgx.CollectExactlyOneRow(rows, pgx.RowToStructByName[sessionModel])
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("collect one row failed: %w", err)
	}

	sess, err := mapSessionToEntity(model)
	if err != nil {
		return nil, err
	}

	return sess, nil
}
