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

type UserRestoreParams struct {
	ID        UserID
	Username  Username
	Status    UserStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

func NewUserWithPassword(
	username Username,
	hashedPassword HashedPassword,
	createdAt time.Time,
) (*User, *UserCredentials, error) {
	if username.IsZero() {
		return nil, nil, ErrUsernameZero
	}
	password, err := NewPasswordCredential(hashedPassword, createdAt)
	if err != nil {
		return nil, nil, err
	}
	userID := NewUserID()
	credentials, err := NewUserCredentialsWithPassword(userID, password)
	if err != nil {
		return nil, nil, err
	}
	user := &User{
		id:        userID,
		username:  username,
		status:    UserStatusActive,
		createdAt: createdAt,
		updatedAt: createdAt,
	}
	return user, credentials, nil
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
func (u User) IsZero() bool {
	return u.id.IsZero() || u.username.IsZero()
}

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
