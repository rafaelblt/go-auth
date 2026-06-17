package credential

import (
	"fmt"
	"strings"
)

type Kind struct {
	value string
}

var (
	KindPassword = Kind{"password"}
)

func ParseKind(value string) (Kind, error) {
	switch strings.TrimSpace(strings.ToLower(value)) {
	case "password":
		return KindPassword, nil
	default:
		return Kind{}, fmt.Errorf("invalid kind: %q", value)
	}
}

func (k Kind) String() string { return k.value }
func (k Kind) IsZero() bool   { return k.value == "" }
