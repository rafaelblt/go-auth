package session

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
)

type RefreshTokenID struct {
	shared.EntityID
}

func NewRefreshTokenID() RefreshTokenID {
	return RefreshTokenID{shared.NewEntityID()}
}

func ParseRefreshTokenID(value string) (RefreshTokenID, error) {
	id, err := shared.ParseEntityID(value)
	if err != nil {
		return RefreshTokenID{}, err
	}
	return RefreshTokenID{id}, nil
}

type RefreshToken struct {
	id        RefreshTokenID
	sessionID SessionID
	hash      RefreshTokenHash
	parentID  *RefreshTokenID
	expiresAt time.Time
	usedAt    *time.Time
	createdAt time.Time
	updatedAt time.Time
}

type RefreshTokenCreationParams struct {
	SessionID SessionID
	Hash      RefreshTokenHash
	ParentID  *RefreshTokenID
	CreatedAt time.Time
	ExpiresAt time.Time
}

type RefreshTokenRestoreParams struct {
	ID        RefreshTokenID
	SessionID SessionID
	Hash      RefreshTokenHash
	ParentID  *RefreshTokenID
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewRefreshToken(params RefreshTokenCreationParams) (*RefreshToken, error) {
	if params.SessionID.IsZero() {
		return nil, errors.New("session id zero")
	}
	if params.Hash.IsZero() {
		return nil, errors.New("hash zero")
	}
	if params.ParentID != nil && params.ParentID.IsZero() {
		return nil, errors.New("parent id zero")
	}
	if params.CreatedAt.After(params.ExpiresAt) {
		return nil, errors.New("created at after expires at")
	}
	token := RefreshToken{
		id:        NewRefreshTokenID(),
		sessionID: params.SessionID,
		hash:      params.Hash,
		parentID:  shared.ClonePtr(params.ParentID),
		expiresAt: params.ExpiresAt,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return &token, nil
}

func RestoreRefreshToken(params RefreshTokenRestoreParams) (*RefreshToken, error) {
	if params.ID.IsZero() {
		return nil, errors.New("id zero")
	}
	if params.SessionID.IsZero() {
		return nil, errors.New("session id zero")
	}
	if params.Hash.IsZero() {
		return nil, errors.New("hash zero")
	}
	if params.ParentID != nil && params.ParentID.IsZero() {
		return nil, errors.New("parent id zero")
	}
	token := RefreshToken{
		id:        params.ID,
		sessionID: params.SessionID,
		hash:      params.Hash,
		parentID:  shared.ClonePtr(params.ParentID),
		expiresAt: params.ExpiresAt,
		usedAt:    shared.ClonePtr(params.UsedAt),
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
	return &token, nil
}

func (t *RefreshToken) ID() RefreshTokenID     { return t.id }
func (t *RefreshToken) SessionID() SessionID   { return t.sessionID }
func (t *RefreshToken) Hash() RefreshTokenHash { return t.hash }
func (t *RefreshToken) ExpiresAt() time.Time   { return t.expiresAt }
func (t *RefreshToken) CreatedAt() time.Time   { return t.createdAt }
func (t *RefreshToken) UpdatedAt() time.Time   { return t.updatedAt }

func (t *RefreshToken) ParentID() (RefreshTokenID, bool) {
	if t.parentID == nil {
		return RefreshTokenID{}, false
	}
	return *t.parentID, true
}

func (t *RefreshToken) UsedAt() (time.Time, bool) {
	if t.usedAt == nil {
		return time.Time{}, false
	}
	return *t.usedAt, true
}

func (t *RefreshToken) IsZero() bool {
	return t.id.IsZero()
}

func (t *RefreshToken) IsUsed() bool {
	return t.usedAt != nil
}

func (t *RefreshToken) HasParent() bool {
	return t.parentID != nil
}

func (t *RefreshToken) Use(usedAt time.Time) error {
	if t.usedAt != nil {
		return ErrTokenAlreadyUsed
	}
	if usedAt.After(t.expiresAt) {
		return ErrTokenExpired
	}
	t.usedAt = &usedAt
	t.updatedAt = usedAt
	return nil
}
