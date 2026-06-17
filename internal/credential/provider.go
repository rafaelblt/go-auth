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
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return Provider{}, errors.New("value empty")
	}
	if normalized == ProviderLocal.value {
		return Provider{}, errors.New("'local' is reserved")
	}
	return Provider{normalized}, nil
}

func ParseProvider(value string) (Provider, error) {
	normalized := strings.ToLower(strings.TrimSpace(value))
	if normalized == "" {
		return Provider{}, errors.New("value empty")
	}
	if normalized == ProviderLocal.value {
		return ProviderLocal, nil
	}
	return Provider{normalized}, nil
}

func (p Provider) String() string { return p.value }
func (p Provider) IsZero() bool   { return p.value == "" }
