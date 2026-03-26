package domain

import (
	"errors"
	"strings"
)

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

func (cp CredentialProvider) String() string { return cp.value }
func (cp CredentialProvider) IsZero() bool   { return cp.value == "" }
