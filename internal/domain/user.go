package domain

import (
	"errors"
	"time"
)

type User struct {
	id        UserID
	username  Username
	status    UserStatus
	createdAt time.Time
	updatedAt time.Time
}

type NewUserParams struct {
	Username  Username
	CreatedAt time.Time
}

type UserRestoreParams struct {
	ID        UserID
	Username  Username
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUser(params NewUserParams) (*User, error) {
	id := NewUserID()
	if params.Username.IsZero() {
		return nil, errors.New("user username cannot be zero")
	}
	user := &User{
		id:        id,
		username:  params.Username,
		status:    UserStatusActive,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return user, nil
}

func RestoreUser(params UserRestoreParams) (*User, error) {
	if params.ID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	if params.Username.IsZero() {
		return nil, errors.New("username cannot be zero")
	}
	return &User{
		id:        params.ID,
		username:  params.Username,
		status:    params.Status,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}, nil
}

func (u User) ID() UserID           { return u.id }
func (u User) Username() Username   { return u.username }
func (u User) Status() UserStatus   { return u.status }
func (u User) CreatedAt() time.Time { return u.createdAt }
func (u User) UpdatedAt() time.Time { return u.updatedAt }
func (u User) IsZero() bool         { return u.id.IsZero() }

func (u *User) ChangeUsername(newUsername Username, updatedAt time.Time) error {
	if newUsername.IsZero() {
		return errors.New("new username cannot be zero")
	}
	if updatedAt.Before(u.createdAt) {
		return errors.New("updated at cannot be before created at")
	}
	u.username = newUsername
	u.updatedAt = updatedAt
	return nil
}
