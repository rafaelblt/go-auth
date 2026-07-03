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

func ParseSessionID(value string) (SessionID, error) {
	id, err := shared.ParseEntityID(value)
	if err != nil {
		return SessionID{}, err
	}
	return SessionID{id}, nil
}

type Session struct {
	id        SessionID
	userID    user.ID
	createdAt time.Time
	revokedAt *time.Time
}

type SessionCreationParams struct {
	UserID    user.ID
	CreatedAt time.Time
}

type SessionRestoreParams struct {
	ID        SessionID
	UserID    user.ID
	CreatedAt time.Time
	RevokedAt *time.Time
}

func NewSession(params SessionCreationParams) (*Session, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	session := Session{
		id:        NewSessionID(),
		userID:    params.UserID,
		createdAt: params.CreatedAt,
	}
	return &session, nil
}

func RestoreSession(params SessionRestoreParams) (*Session, error) {
	if params.ID.IsZero() {
		return nil, errors.New("session id zero")
	}
	if params.UserID.IsZero() {
		return nil, errors.New("user id zero")
	}
	revokedAt := shared.ClonePtr(params.RevokedAt)
	session := Session{
		id:        params.ID,
		userID:    params.UserID,
		createdAt: params.CreatedAt,
		revokedAt: revokedAt,
	}
	return &session, nil
}

func (s Session) ID() SessionID        { return s.id }
func (s Session) UserID() user.ID      { return s.userID }
func (s Session) CreatedAt() time.Time { return s.createdAt }

func (s Session) RevokedAt() (time.Time, bool) {
	if s.revokedAt == nil {
		return time.Time{}, false
	}
	return *s.revokedAt, true
}

func (s Session) IsZero() bool { return s.id.IsZero() }

func (s *Session) IsRevoked() bool {
	return s.revokedAt != nil
}

func (s *Session) Revoke(revokedAt time.Time) {
	if s.revokedAt != nil {
		return
	}
	s.revokedAt = &revokedAt
}
