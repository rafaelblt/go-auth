package credential

import (
	"errors"
	"strings"
)

type Secret struct {
	value string
}

func NewSecret(value string) (Secret, error) {
	normalized := strings.TrimSpace(value)
	if normalized == "" {
		return Secret{}, errors.New("credential secret value cannot be empty")
	}
	return Secret{normalized}, nil
}

func (s Secret) IsZero() bool  { return s.value == "" }
func (s Secret) Value() string { return s.value }
