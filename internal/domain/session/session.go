// Package session holds the Session and RefreshToken entities and their value
// objects. Both live here because they change together: revoking a session
// invalidates its tokens, and reuse detection reads a token and revokes its
// session in one operation.
//
// See docs/architecture/domain/session.md, and docs/architecture/tokens.md for
// what the two token types are for.
package session

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/shared"
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
	revokedAt *time.Time
	createdAt time.Time
	updatedAt time.Time
}

type SessionCreationParams struct {
	UserID    user.ID
	CreatedAt time.Time
}

type SessionRestoreParams struct {
	ID        SessionID
	UserID    user.ID
	RevokedAt *time.Time
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewSession(params SessionCreationParams) (*Session, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	session := Session{
		id:        NewSessionID(),
		userID:    params.UserID,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
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
	session := Session{
		id:        params.ID,
		userID:    params.UserID,
		revokedAt: shared.ClonePtr(params.RevokedAt),
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
	return &session, nil
}

func (s *Session) ID() SessionID        { return s.id }
func (s *Session) UserID() user.ID      { return s.userID }
func (s *Session) CreatedAt() time.Time { return s.createdAt }
func (s *Session) UpdatedAt() time.Time { return s.updatedAt }

func (s *Session) RevokedAt() (time.Time, bool) {
	if s.revokedAt == nil {
		return time.Time{}, false
	}
	return *s.revokedAt, true
}

func (s *Session) IsZero() bool { return s.id.IsZero() }

func (s *Session) IsRevoked() bool {
	return s.revokedAt != nil
}

func (s *Session) Revoke(revokedAt time.Time) {
	if s.revokedAt != nil {
		return
	}
	s.revokedAt = &revokedAt
	s.updatedAt = revokedAt
}
