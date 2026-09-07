package postgres

import (
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/password"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/shared"
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
	Hash      string    `db:"hash"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type sessionModel struct {
	ID        string     `db:"id"`
	UserID    string     `db:"user_id"`
	RevokedAt *time.Time `db:"revoked_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

type refreshTokenModel struct {
	ID        string     `db:"id"`
	SessionID string     `db:"session_id"`
	ParentID  *string    `db:"parent_id"`
	Hash      []byte     `db:"hash"`
	ExpiresAt time.Time  `db:"expires_at"`
	UsedAt    *time.Time `db:"used_at"`
	CreatedAt time.Time  `db:"created_at"`
	UpdatedAt time.Time  `db:"updated_at"`
}

func mapUserToModel(entity *user.User) (userModel, error) {
	if entity == nil {
		return userModel{}, errors.New("user nil")
	}
	if entity.IsZero() {
		return userModel{}, errors.New("user zero")
	}

	model := userModel{
		ID:        entity.ID().String(),
		Username:  entity.Username().String(),
		Status:    entity.Status().String(),
		CreatedAt: entity.CreatedAt(),
		UpdatedAt: entity.UpdatedAt(),
	}

	return model, nil
}

func mapUserToEntity(model userModel) (*user.User, error) {
	id, err := user.ParseID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse user id failed: %w", err)
	}
	username, issues := user.NewUsername(model.Username)
	if !issues.IsEmpty() {
		return nil, fmt.Errorf("username creation failed: %s", issues)
	}
	status, err := user.ParseStatus(model.Status)
	if err != nil {
		return nil, fmt.Errorf("parse user status failed: %w", err)
	}

	usr, err := user.RestoreUser(user.RestoreParams{
		ID:        id,
		Username:  username,
		Status:    status,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore user failed: %w", err)
	}

	return usr, nil
}

func mapCredentialToModel(entity *password.Credential) (credentialModel, error) {
	if entity == nil {
		return credentialModel{}, errors.New("credential nil")
	}
	if entity.IsZero() {
		return credentialModel{}, errors.New("credential zero")
	}

	model := credentialModel{
		ID:        entity.ID().String(),
		UserID:    entity.UserID().String(),
		Hash:      entity.Hash().Value(),
		CreatedAt: entity.CreatedAt(),
		UpdatedAt: entity.UpdatedAt(),
	}

	return model, nil
}

func mapCredentialToEntity(model credentialModel) (*password.Credential, error) {
	id, err := password.ParseID(model.ID)
	if err != nil {
		return nil, fmt.Errorf("parse credential id failed: %w", err)
	}
	userID, err := user.ParseID(model.UserID)
	if err != nil {
		return nil, fmt.Errorf("parse user id failed: %w", err)
	}
	hash, err := password.NewHashed(model.Hash)
	if err != nil {
		return nil, fmt.Errorf("new hashed password failed: %w", err)
	}

	cred, err := password.RestoreCredential(password.RestoreParams{
		ID:        id,
		UserID:    userID,
		Hash:      hash,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore credential failed: %w", err)
	}

	return cred, nil
}

func mapSessionToModel(entity *session.Session) (sessionModel, error) {
	if entity == nil {
		return sessionModel{}, errors.New("session nil")
	}
	if entity.IsZero() {
		return sessionModel{}, errors.New("session zero")
	}

	model := sessionModel{
		ID:        entity.ID().String(),
		UserID:    entity.UserID().String(),
		RevokedAt: shared.PtrFromOk(entity.RevokedAt()),
		CreatedAt: entity.CreatedAt(),
		UpdatedAt: entity.UpdatedAt(),
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
		RevokedAt: model.RevokedAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore session failed: %w", err)
	}

	return sess, nil
}

func mapRefreshTokenToModel(entity *session.RefreshToken) (refreshTokenModel, error) {
	if entity == nil {
		return refreshTokenModel{}, errors.New("refresh token nil")
	}
	if entity.IsZero() {
		return refreshTokenModel{}, errors.New("refresh token zero")
	}

	var parentID *string
	pID, ok := entity.ParentID()
	if ok {
		id := pID.String()
		parentID = &id
	}

	model := refreshTokenModel{
		ID:        entity.ID().String(),
		SessionID: entity.SessionID().String(),
		ParentID:  parentID,
		Hash:      entity.Hash().Value(),
		ExpiresAt: entity.ExpiresAt(),
		UsedAt:    shared.PtrFromOk(entity.UsedAt()),
		CreatedAt: entity.CreatedAt(),
		UpdatedAt: entity.UpdatedAt(),
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
		ExpiresAt: model.ExpiresAt,
		UsedAt:    model.UsedAt,
		CreatedAt: model.CreatedAt,
		UpdatedAt: model.UpdatedAt,
	})
	if err != nil {
		return nil, fmt.Errorf("restore refresh token failed: %w", err)
	}

	return token, nil
}
