package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
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
	model, err := repo.mapToModel(sess)
	if err != nil {
		return fmt.Errorf("map session to model failed: %w", err)
	}

	sql := `INSERT INTO sessions
			(id, user_id, issued_at, revoked_at)
			VALUES ($1, $2, $3, $4)`
	_, err = repo.db.Exec(ctx, sql,
		model.ID,
		model.UserID,
		model.IssuedAt,
		model.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("session insert failed: %w", err)
	}

	return nil
}

func (repo *SessionRepo) Update(ctx context.Context, sess *session.Session) error {
	model, err := repo.mapToModel(sess)
	if err != nil {
		return fmt.Errorf("map session to model failed: %w", err)
	}

	sql := `UPDATE sessions
			SET revoked_at = $2
			WHERE id = $1`
	tag, err := repo.db.Exec(ctx, sql,
		model.ID,
		model.RevokedAt,
	)

	if err != nil {
		return fmt.Errorf("session update failed: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("no rows affected in session update")
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

	sess, err := repo.mapToEntity(model)
	if err != nil {
		return nil, err
	}

	return sess, nil
}

func (repo *SessionRepo) mapToModel(sess *session.Session) (sessionModel, error) {
	if sess == nil {
		return sessionModel{}, errors.New("session nil")
	}
	if sess.IsZero() {
		return sessionModel{}, errors.New("session zero")
	}

	id := sess.ID().String()
	userID := sess.UserID().String()
	issuedAt := sess.IssuedAt()
	var revokedAtPtr *time.Time

	revokedAt, ok := sess.RevokedAt()
	if ok {
		revokedAtPtr = &revokedAt
	}

	model := sessionModel{
		model:     model{ID: id},
		UserID:    userID,
		IssuedAt:  issuedAt,
		RevokedAt: revokedAtPtr,
	}

	return model, nil
}

func (repo *SessionRepo) mapToEntity(model sessionModel) (*session.Session, error) {
	id, err := session.ParseSessionID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse session id failed: %w", err)
	}
	userID, err := user.ParseID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id failed: %w", err)
	}

	sess, err := session.RestoreSession(session.SessionRestoreParams{
		ID:        id,
		UserID:    userID,
		IssuedAt:  model.IssuedAt,
		RevokedAt: model.RevokedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore session failed: %w", err)
	}

	return sess, nil
}
