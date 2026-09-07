// Package password holds the credential a user authenticates with:
// a plain password supplied at register/login time, and the hash of it
// that is the only form ever persisted.
package password

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
)

type Credential struct {
	id        ID
	userID    user.ID
	hash      Hashed
	createdAt time.Time
	updatedAt time.Time
}

type CreationParams struct {
	UserID    user.ID
	Hash      Hashed
	CreatedAt time.Time
}

func NewCredential(params CreationParams) (*Credential, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	if params.Hash.IsZero() {
		return nil, errors.New("hash cannot be zero")
	}
	cred := &Credential{
		id:        NewID(),
		userID:    params.UserID,
		hash:      params.Hash,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return cred, nil
}

type RestoreParams struct {
	ID        ID
	UserID    user.ID
	Hash      Hashed
	CreatedAt time.Time
	UpdatedAt time.Time
}

func RestoreCredential(params RestoreParams) (*Credential, error) {
	if params.ID.IsZero() {
		return nil, errors.New("id cannot be zero")
	}
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	if params.Hash.IsZero() {
		return nil, errors.New("hash cannot be zero")
	}
	if params.CreatedAt.IsZero() {
		return nil, errors.New("created at cannot be zero")
	}
	if params.UpdatedAt.IsZero() {
		return nil, errors.New("updated at cannot be zero")
	}
	cred := &Credential{
		id:        params.ID,
		userID:    params.UserID,
		hash:      params.Hash,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
	return cred, nil
}

func (c *Credential) ID() ID               { return c.id }
func (c *Credential) UserID() user.ID      { return c.userID }
func (c *Credential) Hash() Hashed         { return c.hash }
func (c *Credential) CreatedAt() time.Time { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time { return c.updatedAt }

func (c *Credential) IsZero() bool { return c.id.IsZero() }
