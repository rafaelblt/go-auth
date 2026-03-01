package domain

import (
	"github.com/google/uuid"
)

type EntityID struct {
	value uuid.UUID
}

type UserID struct {
	EntityID
}


func newEntityID() EntityID {
	value := uuid.New()
	return EntityID{value: value}
}

func (id EntityID) Value() uuid.UUID { return id.value }

func NewUserID() UserID { return UserID{newEntityID()} }
