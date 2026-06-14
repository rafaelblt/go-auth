package session

import (
	"errors"
	"slices"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
)

type RefreshTokenID struct {
	shared.EntityID
}

func NewRefreshTokenID() RefreshTokenID {
	return RefreshTokenID{shared.NewEntityID()}
}

type RefreshToken struct {
	id        RefreshTokenID
	sessionID SessionID
	hash      []byte
	parentID  *RefreshTokenID
	issuedAt  time.Time
	expiresAt time.Time
	usedAt    *time.Time
}

type RefreshTokenCreationParams struct {
	SessionID SessionID
	Hash      []byte
	ParentID  *RefreshTokenID
	IssuedAt  time.Time
	ExpiresAt time.Time
}

func NewRefreshToken(params RefreshTokenCreationParams) (*RefreshToken, error) {
	if params.SessionID.IsZero() {
		return nil, errors.New("session id zero")
	}
	if params.Hash == nil {
		return nil, errors.New("hash nil")
	}
	if len(params.Hash) == 0 {
		return nil, errors.New("hash empty")
	}
	if params.ParentID != nil && params.ParentID.IsZero() {
		return nil, errors.New("parent id zero")
	}
	if params.IssuedAt.After(params.ExpiresAt) {
		return nil, errors.New("issued at after expires at")
	}
	token := RefreshToken{
		id:        NewRefreshTokenID(),
		sessionID: params.SessionID,
		hash:      slices.Clone(params.Hash),
		parentID:  shared.ClonePtr(params.ParentID),
		issuedAt:  params.IssuedAt,
		expiresAt: params.ExpiresAt,
	}
	return &token, nil
}

func (t RefreshToken) ID() RefreshTokenID        { return t.id }
func (t RefreshToken) SessionID() SessionID      { return t.sessionID }
func (t RefreshToken) Hash() []byte              { return slices.Clone(t.hash) }
func (t RefreshToken) IssuedAt() time.Time       { return t.issuedAt }
func (t RefreshToken) ExpiresAt() time.Time      { return t.expiresAt }

func (t RefreshToken) ParentID() (RefreshTokenID, bool) {
	if t.parentID == nil {
		return RefreshTokenID{}, false
	}
	return *t.parentID, true
}

func (t RefreshToken) UsedAt() (time.Time, bool) {
	if t.usedAt == nil {
		return time.Time{}, false
	}
	return *t.usedAt, true
}

func (t RefreshToken) IsZero() bool { return t.id.IsZero() }

func (t RefreshToken) HasParent() bool {
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
	return nil
}
