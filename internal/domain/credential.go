package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// Credential

type Credential struct {
	id        CredentialID
	userID    UserID
	kind      CredentialKind
	provider  CredentialProvider
	secret    CredentialSecret
	createdAt time.Time
	updatedAt time.Time
}

type NewCredentialParams struct {
	UserID    UserID
	Kind      CredentialKind
	Provider  CredentialProvider
	Secret    CredentialSecret
	CreatedAt time.Time
}

func NewCredential(params NewCredentialParams) (*Credential, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("credential user id cannot be zero")
	}
	if !params.Kind.IsValid() {
		return nil, fmt.Errorf("credential kind '%v' is not valid", params.Kind)
	}
	if params.Provider.IsZero() {
		return nil, errors.New("credential provider cannot be zero")
	}
	if params.Secret.IsZero() {
		return nil, errors.New("credential secret cannot be zero")
	}
	cred := &Credential{
		id:        NewCredentialID(),
		userID:    params.UserID,
		kind:      params.Kind,
		provider:  params.Provider,
		secret:    params.Secret,
		createdAt: params.CreatedAt,
		updatedAt: params.CreatedAt,
	}
	return cred, nil
}

func (c Credential) ID() CredentialID             { return c.id }
func (c Credential) UserID() UserID               { return c.userID }
func (c Credential) Kind() CredentialKind         { return c.kind }
func (c Credential) Provider() CredentialProvider { return c.provider }
func (c Credential) Secret() CredentialSecret     { return c.secret }
func (c Credential) CreatedAt() time.Time         { return c.createdAt }
func (c Credential) UpdatedAt() time.Time         { return c.updatedAt }

func (c Credential) IsZero() bool { return c.id.IsZero() }

// Credential Kind

type CredentialKind string

const (
	CredentialKindPassword CredentialKind = "password"
)

func (k CredentialKind) String() string { return string(k) }
func (k CredentialKind) IsValid() bool {
	switch k {
	case CredentialKindPassword:
		return true
	default:
		return false
	}
}

// Credential Provider

type CredentialProvider struct {
	value string
}

var CredentialProviderLocal = CredentialProvider{value: "local"}

func NewCredentialProvider(value string) (CredentialProvider, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return CredentialProvider{}, errors.New("credential provider value cannot be empty")
	}
	if normalized == CredentialProviderLocal.value {
		return CredentialProvider{}, errors.New("credential provider 'local' is reserved")
	}
	return CredentialProvider{normalized}, nil
}

func (cp CredentialProvider) IsZero() bool   { return cp.value == "" }
func (cp CredentialProvider) String() string { return cp.value }

// Credential Secret

type CredentialSecret struct {
	value string
}

func NewCredentialSecret(value string) (CredentialSecret, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return CredentialSecret{}, errors.New("credential secret value cannot be empty")
	}
	return CredentialSecret{normalized}, nil
}

func (cs CredentialSecret) IsZero() bool  { return cs.value == "" }
func (cs CredentialSecret) Value() string { return cs.value }
