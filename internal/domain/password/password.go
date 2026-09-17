package password

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/user"
)

type Password struct {
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

func NewPassword(params CreationParams) (*Password, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	if params.Hash.IsZero() {
		return nil, errors.New("hash cannot be zero")
	}
	pwd := &Password{
		id:        NewID(),
		userID:    params.UserID,
		hash:      params.Hash,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return pwd, nil
}

type RestoreParams struct {
	ID        ID
	UserID    user.ID
	Hash      Hashed
	CreatedAt time.Time
	UpdatedAt time.Time
}

func RestorePassword(params RestoreParams) (*Password, error) {
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
	pwd := &Password{
		id:        params.ID,
		userID:    params.UserID,
		hash:      params.Hash,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
	return pwd, nil
}

func (p *Password) ID() ID               { return p.id }
func (p *Password) UserID() user.ID      { return p.userID }
func (p *Password) Hash() Hashed         { return p.hash }
func (p *Password) CreatedAt() time.Time { return p.createdAt }
func (p *Password) UpdatedAt() time.Time { return p.updatedAt }

func (p *Password) IsZero() bool { return p.id.IsZero() }
