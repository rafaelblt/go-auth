package user

import (
	"strings"

	"github.com/rafaelblt/go-auth/internal/validation"
)

type Username struct {
	value string
}

const (
	UsernameMinLen = 3
	UsernameMaxLen = 32
)

func NewUsername(value string) (Username, error) {
	normalized := normalizeUsername(value)

	issues := make(validation.Issues, 0)
	if len(normalized) < UsernameMinLen {
		issues = append(issues, validation.IssueMinLen(UsernameMinLen))
	}
	if len(normalized) > UsernameMaxLen {
		issues = append(issues, validation.IssueMaxLen(UsernameMaxLen))
	}

	if len(issues) > 0 {
		return Username{}, issues
	}

	return Username{normalized}, nil
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (u Username) IsZero() bool   { return u.value == "" }
func (u Username) String() string { return u.value }
