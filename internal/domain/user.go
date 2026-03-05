package domain

import (
	"time"
)

type User struct {
	id        UserID
	username  Username
	status    UserStatus
	createdAt time.Time
	updatedAt time.Time
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
	user :=  &User{
		id:        userID,
		username:  username,
		status:    UserStatusActive,
		createdAt: createdAt,
		updatedAt: createdAt,
	}
	return user, credentials, nil
}

func (u User) ID() UserID           { return u.id }
func (u User) Username() Username   { return u.username }
func (u User) Status() UserStatus   { return u.status }
func (u User) CreatedAt() time.Time { return u.createdAt }
func (u User) UpdatedAt() time.Time { return u.updatedAt }
