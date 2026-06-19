package shared

import (
	"errors"

	"github.com/google/uuid"
)

type EntityID struct {
	value uuid.UUID
}

func NewEntityID() EntityID {
	value := uuid.New()
	return EntityID{value: value}
}

func newEntityIDFrom(id uuid.UUID) (EntityID, error) {
	if id == uuid.Nil {
		return EntityID{}, errors.New("uuid cannot be nil")
	}
	return EntityID{id}, nil
}

func ParseEntityID(value string) (EntityID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil {
		return EntityID{}, err
	}
	id, err := newEntityIDFrom(parsed)
	if err != nil {
		return EntityID{}, err
	}
	return id, nil
}

func (id EntityID) Value() uuid.UUID { return id.value }
func (id EntityID) String() string   { return id.value.String() }
func (id EntityID) IsZero() bool     { return id.value == uuid.Nil }
