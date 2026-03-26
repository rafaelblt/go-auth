package domain

import (
	"errors"
	"strings"
)

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
