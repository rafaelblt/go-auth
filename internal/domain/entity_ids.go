package domain

import (
	"errors"

	"github.com/google/uuid"
)

type EntityID struct {
	value uuid.UUID
}

type UserID struct {
	EntityID
}

type CredentialID struct {
	EntityID
}

func newEntityID() EntityID {
	value := uuid.New()
	return EntityID{value: value}
}
func newEntityIDFrom(id uuid.UUID) (EntityID, error) {
	if id == uuid.Nil {
		return EntityID{}, errors.New("uuid cannot be nil")
	}
	return EntityID{id}, nil
}

func (id EntityID) Value() uuid.UUID { return id.value }
func (id EntityID) IsZero() bool { return id.value == uuid.Nil }

func NewUserID() UserID { return UserID{newEntityID()} }
func NewCredentialID() CredentialID { return CredentialID{newEntityID()} }

func NewUserIDFrom(id uuid.UUID) (UserID, error) {
	converted, err := newEntityIDFrom(id)
	if err != nil {
		return UserID{}, err
	}
	return UserID{converted}, nil
}
