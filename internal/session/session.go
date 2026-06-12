package session

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/user"
)

type SessionID struct {
	shared.EntityID
}

func NewSessionID() SessionID {
	return SessionID{shared.NewEntityID()}
}

type Session struct {
	id        SessionID
	userID    user.ID
	issuedAt  time.Time
	revokedAt *time.Time
}

type SessionCreationParams struct {
	UserID   user.ID
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
func (s Session) UserID() user.ID     { return s.userID }
func (s Session) IssuedAt() time.Time { return s.issuedAt }

func (s Session) RevokedAt() (time.Time, bool)  {
	if s.revokedAt == nil {
		return time.Time{}, false
	}
	return *s.revokedAt, true
}

func (s Session) IsZero() bool { return s.id.IsZero() }
