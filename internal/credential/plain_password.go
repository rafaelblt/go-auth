package credential

import (
	"strings"

	"github.com/rafaelblt/go-auth/internal/validation"
)

type PlainPassword struct {
	value string
}

const (
	PlainPasswordMaxLen = 32
	PlainPasswordMinLen = 4
)

func NewPlainPassword(value string) (PlainPassword, error) {
	normalized := normalizePlainPassword(value)

	issues := make(validation.Issues, 0)
	if len(normalized) < PlainPasswordMinLen {
		issues = append(issues, validation.IssueMinLen(PlainPasswordMinLen))
	}
	if len(normalized) > PlainPasswordMaxLen {
		issues = append(issues, validation.IssueMaxLen(PlainPasswordMaxLen))
	}

	if len(issues) > 0 {
		return PlainPassword{}, issues
	}

	return PlainPassword{value: normalized}, nil
}

func normalizePlainPassword(value string) string {
	return strings.TrimSpace(value)
}

func (p PlainPassword) Value() string { return p.value }
func (p PlainPassword) IsZero() bool  { return p.value == "" }
