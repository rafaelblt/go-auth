package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
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

func (repo *SessionRepo) Save(ctx context.Context, sess *session.Session) error {
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
