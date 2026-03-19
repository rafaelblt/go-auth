package domain

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

type AuthCredential struct {
	id        CredentialID
	userID    UserID
	kind      CredentialKind
	provider  CredentialProvider
	secret    CredentialSecret
	createdAt time.Time
	updatedAt time.Time
}

type CredentialKind string

type CredentialProvider struct {
	value string
}

type CredentialSecret struct {
	value string
}
type AuthCredentialsParams struct {
	UserID    UserID
	Kind      CredentialKind
	Provider  CredentialProvider
	Secret    CredentialSecret
	CreatedAt time.Time
}

const (
	CredentialKindPassword CredentialKind = "password"
)
var CredentialProviderLocal = CredentialProvider{value: "local"}

func NewAuthCredential(params AuthCredentialsParams) (*AuthCredential, error) {
	if params.UserID.IsZero() {
		return nil, errors.New("auth credential user id cannot be zero")
	}
	if !params.Kind.IsValid() {
		return nil, fmt.Errorf("auth credential kind '%v' is not valid", params.Kind)
	}
	if params.Provider.IsZero() {
		return nil, errors.New("auth credential provider cannot be zero")
	}
	if params.Secret.IsZero() {
		return nil, errors.New("auth credential secret cannot be zero")
	}
	cred := &AuthCredential{
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

func NewCredentialSecret(value string) (CredentialSecret, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return CredentialSecret{}, errors.New("credential secret value cannot be empty")
	}
	return CredentialSecret{normalized}, nil
}

func (ac AuthCredential) ID() CredentialID             { return ac.id }
func (ac AuthCredential) UserID() UserID               { return ac.userID }
func (ac AuthCredential) Kind() CredentialKind         { return ac.kind }
func (ac AuthCredential) Provider() CredentialProvider { return ac.provider }
func (ac AuthCredential) Secret() CredentialSecret     { return ac.secret }
func (ac AuthCredential) CreatedAt() time.Time         { return ac.createdAt }
func (ac AuthCredential) UpdatedAt() time.Time         { return ac.updatedAt }

func (k CredentialKind) String() string { return string(k) }
func (k CredentialKind) IsValid() bool {
	switch k {
	case CredentialKindPassword:
		return true
	default:
		return false
	}
}

func (cp CredentialProvider) IsZero() bool { return cp.value == "" }
func (cp CredentialProvider) String() string { return cp.value }

func (cs CredentialSecret) IsZero() bool { return cs.value == "" }
func (cs CredentialSecret) Value() string { return cs.value }
