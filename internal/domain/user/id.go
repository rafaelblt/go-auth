package user

import "github.com/rafaelblt/go-auth/internal/shared"

type ID struct {
	shared.EntityID
}

func NewID() ID {
	return ID{shared.NewEntityID()}
}

func ParseID(value string) (ID, error) {
	id, err := shared.ParseEntityID(value)
	if err != nil {
		return ID{}, err
	}
	return ID{id}, nil
}
