package infra

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/credential"
	"github.com/rafaelblt/go-auth/internal/user"
)

func MapUserToModel(usr *user.User) (UserModel, error) {
	if usr == nil {
		return UserModel{}, errors.New("user cannot be nil to map to model")
	}
	if usr.IsZero() {
		return UserModel{}, errors.New("user cannot be zero to map to model")
	}
	idBytes, err := usr.ID().Value().MarshalBinary()
	if err != nil {
		return UserModel{}, fmt.Errorf("uuid conversion failed: %w", err)
	}
	model := UserModel{
		ID:        idBytes,
		Username:  usr.Username().String(),
		Status:    usr.Status().String(),
		CreatedAt: usr.CreatedAt(),
		UpdatedAt: usr.UpdatedAt(),
	}
	return model, nil
}

func MapCredentialToModel(cred *credential.Credential) (CredentialModel, error) {
	if cred == nil {
		return CredentialModel{}, errors.New("credential cannot be nil to map to model")
	}
	if cred.IsZero() {
		return CredentialModel{}, errors.New("credential cannot be zero to map to model")
	}
	idBytes, err := cred.ID().Value().MarshalBinary()
	if err != nil {
		return CredentialModel{}, fmt.Errorf("credential id to uuid conversion failed: %w", err)
	}
	userIDBytes, err := cred.UserID().Value().MarshalBinary()
	if err != nil {
		return CredentialModel{}, fmt.Errorf("credential user id to uuid conversion failed: %w", err)
	}
	model := CredentialModel{
		ID:        idBytes,
		UserID:    userIDBytes,
		Kind:      cred.Kind().String(),
		Provider:  cred.Provider().String(),
		Secret:    cred.Secret().Value(),
		CreatedAt: cred.CreatedAt(),
		UpdatedAt: cred.UpdatedAt(),
	}
	return model, nil
}
