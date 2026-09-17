package user

import (
	"errors"
	"time"
)

type User struct {
	id        ID
	username  Username
	status    Status
	createdAt time.Time
	updatedAt time.Time
}

type CreationParams struct {
	Username  Username
	CreatedAt time.Time
}

type RestoreParams struct {
	ID        ID
	Username  Username
	Status    Status
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(params CreationParams) (*User, error) {
	id := NewID()
	if params.Username.IsZero() {
		return nil, errors.New("username zero")
	}
	user := &User{
		id:        id,
		username:  params.Username,
		status:    StatusActive,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return user, nil
}

func RestoreUser(params RestoreParams) (*User, error) {
	if params.ID.IsZero() {
		return nil, errors.New("user id zero")
	}
	if params.Username.IsZero() {
		return nil, errors.New("username zero")
	}
	if params.Status.IsZero() {
		return nil, errors.New("status zero")
	}
	return &User{
		id:        params.ID,
		username:  params.Username,
		status:    params.Status,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}, nil
}

func (u *User) ID() ID               { return u.id }
func (u *User) Username() Username   { return u.username }
func (u *User) Status() Status       { return u.status }
func (u *User) CreatedAt() time.Time { return u.createdAt }
func (u *User) UpdatedAt() time.Time { return u.updatedAt }
func (u *User) IsZero() bool         { return u.id.IsZero() }
