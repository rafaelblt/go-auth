package domain

import (
	"time"
)

type PasswordCredential struct {
	id        CredentialID
	hashedPwd HashedPassword
	createdAt time.Time
	updatedAt time.Time
}

func NewPasswordCredential(
	hashedPwd HashedPassword,
	createdAt time.Time,
) (PasswordCredential, error) {
	if hashedPwd.IsZero() {
		return PasswordCredential{}, ErrHashedPasswordZero
	}
	return PasswordCredential{
		id:        NewCredentialID(),
		hashedPwd: hashedPwd,
		createdAt: createdAt,
		updatedAt: createdAt,
	}, nil
}
func (p PasswordCredential) ID() CredentialID       { return p.id }
func (p PasswordCredential) Hashed() HashedPassword { return p.hashedPwd }
func (p PasswordCredential) CreatedAt() time.Time   { return p.createdAt }
func (p PasswordCredential) UpdatedAt() time.Time   { return p.updatedAt }
func (p PasswordCredential) IsZero() bool           { return p.hashedPwd.IsZero() }
