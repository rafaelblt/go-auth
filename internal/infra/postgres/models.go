package postgres

import (
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/user"
)

type Model struct {
	ID string `db:"id"`
}

type userModel struct {
	ID        string    `db:"id"`
	Username  string    `db:"username"`
	Status    string    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type credentialModel struct {
	ID        string    `db:"id"`
	UserID    string    `db:"user_id"`
	Kind      string    `db:"kind"`
	Provider  string    `db:"provider"`
	Secret    string    `db:"secret"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sessionModel struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	CreatedAt time.Time  `db:"created_at"`
	RevokedAt *time.Time `db:"revoked_at"`
}

type refreshTokenModel struct {
	ID        string     `db:"id"`
	SessionID string     `db:"session_id"`
	ParentID  *string    `db:"parent_id"`
	Hash      []byte     `db:"hash"`
	CreatedAt time.Time  `db:"created_at"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
}

func mapSessionToModel(sess *session.Session) (sessionModel, error) {
	if sess == nil {
		return sessionModel{}, errors.New("session nil")
	}
	if sess.IsZero() {
		return sessionModel{}, errors.New("session zero")
	}

	id := sess.ID().String()
	userID := sess.UserID().String()
	createdAt := sess.CreatedAt()
	var revokedAtPtr *time.Time

	revokedAt, ok := sess.RevokedAt()
	if ok {
		revokedAtPtr = &revokedAt
	}

	model := sessionModel{
		ID:        id,
		UserID:    userID,
		CreatedAt: createdAt,
		RevokedAt: revokedAtPtr,
	}

	return model, nil
}

func mapSessionToEntity(model sessionModel) (*session.Session, error) {
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
		CreatedAt: model.CreatedAt,
		RevokedAt: model.RevokedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore session failed: %w", err)
	}

	return sess, nil
}

func mapRefreshTokenToModel(token *session.RefreshToken) (refreshTokenModel, error) {
	if token == nil {
		return refreshTokenModel{}, errors.New("refresh token nil")
	}
	if token.IsZero() {
		return refreshTokenModel{}, errors.New("refresh token zero")
	}

	id := token.ID().String()
	sessionID := token.SessionID().String()
	hash := token.Hash().Value()
	createdAt := token.CreatedAt()
	expiresAt := token.ExpiresAt()

	var parentID *string
	var usedAt *time.Time

	pID, ok := token.ParentID()
	if ok {
		id := pID.String()
		parentID = &id
	}

	uAt, ok := token.UsedAt()
	if ok {
		usedAt = &uAt
	}

	model := refreshTokenModel{
		ID:        id,
		SessionID: sessionID,
		ParentID:  parentID,
		Hash:      hash,
		CreatedAt: createdAt,
		ExpiresAt: expiresAt,
		UsedAt:    usedAt,
	}

	return model, nil
}

func mapRefreshTokenToEntity(model refreshTokenModel) (*session.RefreshToken, error) {
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
		CreatedAt: model.CreatedAt,
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore refresh token failed: %w", err)
	}

	return token, nil
}
