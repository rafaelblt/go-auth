package infra

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/domain"
)

func MapUserDomainToModel(user *domain.User) (UserModel, error) {
	if user == nil {
		return UserModel{}, errors.New("user cannot be nil to map to model")
	}
	if user.IsZero() {
		return UserModel{}, errors.New("user cannot be zero to map to model")
	}
	idBytes, err := user.ID().Value().MarshalBinary()
	if err != nil {
		return UserModel{}, fmt.Errorf("uuid conversion failed: %w", err)
	}
	model := UserModel{
		ID:        idBytes,
		Username:  user.Username().String(),
		Status:    user.Status().String(),
		CreatedAt: user.CreatedAt(),
		UpdatedAt: user.UpdatedAt(),
	}
	return model, nil
}
