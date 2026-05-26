package credential

import (
	"errors"
	"strings"
)

type Provider struct {
	value string
}

var ProviderLocal = Provider{value: "local"}

func NewProvider(value string) (Provider, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return Provider{}, errors.New("credential provider value cannot be empty")
	}
	if normalized == ProviderLocal.value {
		return Provider{}, errors.New("credential provider 'local' is reserved")
	}
	return Provider{normalized}, nil
}

func (cp Provider) String() string { return cp.value }
func (cp Provider) IsZero() bool   { return cp.value == "" }
