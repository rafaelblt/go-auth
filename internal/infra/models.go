package infra

import (
	"time"
)

type UserModel struct {
	ID        []byte
	Username  string
	Status    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m UserModel) IsZero() bool { return m.ID == nil }

type CredentialModel struct {
	ID        []byte
	UserID    []byte
	Kind      string
	Provider  string
	Secret    string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (m CredentialModel) IsZero() bool { return m.ID == nil }
