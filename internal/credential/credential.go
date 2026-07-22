package credential

import (
	"errors"
	"time"

	"github.com/rafaelblt/go-auth/internal/user"
)

type Credential struct {
	id        ID
	userID    user.ID
	kind      Kind
	provider  Provider
	secret    Secret
	createdAt time.Time
	updatedAt time.Time
}

type CreationParams struct {
	UserID    user.ID
	Kind      Kind
	Provider  Provider
	Secret    Secret
	CreatedAt time.Time
}

func NewCredential(params CreationParams) (*Credential, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("user id cannot be zero")
	}
	if params.Kind.IsZero() {
		return nil, errors.New("kind cannot be zero")
	}
	if params.Provider.IsZero() {
		return nil, errors.New("provider cannot be zero")
	}
	if params.Secret.IsZero() {
		return nil, errors.New("secret cannot be zero")
	}
	cred := &Credential{
		id:        NewID(),
		userID:    params.UserID,
		kind:      params.Kind,
		provider:  params.Provider,
		secret:    params.Secret,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return cred, nil
}

type RestoreParams struct {
	ID        ID
	UserID    user.ID
	Kind      Kind
	Provider  Provider
	Secret    Secret
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
	if params.Kind.IsZero() {
		return nil, errors.New("kind cannot be zero")
	}
	if params.Provider.IsZero() {
		return nil, errors.New("provider cannot be zero")
	}
	if params.Secret.IsZero() {
		return nil, errors.New("secret cannot be zero")
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
		kind:      params.Kind,
		provider:  params.Provider,
		secret:    params.Secret,
		createdAt: params.CreatedAt,
		updatedAt: params.UpdatedAt,
	}
	return cred, nil
}

func (c *Credential) ID() ID               { return c.id }
func (c *Credential) UserID() user.ID      { return c.userID }
func (c *Credential) Kind() Kind           { return c.kind }
func (c *Credential) Provider() Provider   { return c.provider }
func (c *Credential) Secret() Secret       { return c.secret }
func (c *Credential) CreatedAt() time.Time { return c.createdAt }
func (c *Credential) UpdatedAt() time.Time { return c.updatedAt }

func (c *Credential) IsZero() bool { return c.id.IsZero() }
