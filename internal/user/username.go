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

	normalized := normalizeUsername(value)

	issues := make(validation.Issues, 0)
	if len(normalized) < UsernameMinLen {
		issues = append(issues, validation.IssueTooShort(UsernameMinLen))
	}
	if len(normalized) > UsernameMaxLen {
		issues = append(issues, validation.IssueTooLong(UsernameMaxLen))
	}
func NewUsername(value string) (Username, validation.Issues) {

	if !issues.IsEmpty() {
		return Username{}, issues
	}

	return Username{normalized}, nil
}

func normalizeUsername(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (u Username) IsZero() bool   { return u.value == "" }
func (u Username) String() string { return u.value }
