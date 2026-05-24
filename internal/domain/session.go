package domain

import (
	"errors"
	"time"
)

type SessionID struct {
	EntityID
}

func NewSessionID() SessionID {
	return SessionID{newEntityID()}
}

type Session struct {
	id        SessionID
	userID    UserID
	issuedAt  time.Time
	revokedAt *time.Time
}

type SessionCreationParams struct {
	UserID   UserID
	IssuedAt time.Time
}

func NewSession(params SessionCreationParams) (*Session, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	session := Session{
		id:       NewSessionID(),
		userID:   params.UserID,
		issuedAt: params.IssuedAt,
	}
	return &session, nil
}

func (s Session) ID() SessionID       { return s.id }
func (s Session) UserID() UserID      { return s.userID }
func (s Session) IssuedAt() time.Time { return s.issuedAt }
func (s Session) UsedAt() *time.Time  { return s.revokedAt }

func (s Session) IsZero() bool { return s.id.IsZero() }
