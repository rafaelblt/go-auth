package domain

import (
	"errors"
	"time"
)

type RefreshTokenID struct {
	EntityID
}

func NewRefreshTokenID() RefreshTokenID {
	return RefreshTokenID{newEntityID()}
}

type RefreshToken struct {
	id        RefreshTokenID
	sessionID SessionID
	hash      string
	parentID  *RefreshTokenID
	issuedAt  time.Time
	expiresAt time.Time
	usedAt    *time.Time
}

type RefreshTokenCreationParams struct {
	SessionID SessionID
	Hash      string
	ParentID  *RefreshTokenID
	IssuedAt  time.Time
	ExpiresAt time.Time
	UsedAt    *time.Time
}

func NewRefreshToken(params RefreshTokenCreationParams) (*RefreshToken, error) {
	if params.SessionID.IsZero() {
		return nil, errors.New("session id cannot be zero")
	}
	if params.Hash == "" {
		return nil, errors.New("hash cannot be empty")
	}
	if params.ParentID != nil && params.ParentID.IsZero() {
		return nil, errors.New("parent id cannot be zero")
	}
	token := RefreshToken{
		id:        NewRefreshTokenID(),
		sessionID: params.SessionID,
		hash:      params.Hash,
		parentID:  params.ParentID,
		issuedAt:  params.IssuedAt,
		expiresAt: params.ExpiresAt,
		usedAt:    params.UsedAt,
	}
	return &token, nil
}

func (t RefreshToken) ID() RefreshTokenID        { return t.id }
func (t RefreshToken) SessionID() SessionID      { return t.sessionID }
func (t RefreshToken) Hash() string              { return t.hash }
func (t RefreshToken) ParentID() *RefreshTokenID { return t.parentID }
func (t RefreshToken) IssuedAt() time.Time       { return t.issuedAt }
func (t RefreshToken) ExpiresAt() time.Time      { return t.expiresAt }
func (t RefreshToken) UsedAt() *time.Time        { return t.usedAt }

func (t RefreshToken) IsZero() bool { return t.id.IsZero() }
