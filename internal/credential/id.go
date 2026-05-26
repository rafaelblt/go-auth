package credential

import "github.com/rafaelblt/go-auth/internal/shared"

type ID struct {
	shared.EntityID
}

func NewID() ID {
	return ID{shared.NewEntityID()}
}
