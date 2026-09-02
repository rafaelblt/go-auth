package user

import (
	"strings"

	"github.com/rafaelblt/go-auth/internal/shared"
	"github.com/rafaelblt/go-auth/internal/validation"
)

type Username struct {
	value string
}

const (
	UsernameMinLen = 3
	UsernameMaxLen = 32
)

var usernameAllowedChars = buildUsernameAllowedChars(
	"abcdefghijklmnopqrstuvwxyz0123456789._-",
)

func buildUsernameAllowedChars(s string) shared.Set[rune] {
	set := shared.NewSet[rune]()
	for _, r := range s {
		set.Add(r)
	}
	return set
}

func NewUsername(value string) (Username, validation.Issues) {
	normalized := normalizeUsername(value)

	issues := validation.Validate(normalized,
		validation.MinLength(UsernameMinLen),
		validation.MaxLength(UsernameMaxLen),
		validation.AllowedChars(usernameAllowedChars),
	)

	if !issues.IsEmpty() {
		return Username{}, issues
	}

	return Username{normalized}, nil
}

func normalizeUsername(value string) string {
	return strings.ToLower(value)
}

func (u Username) IsZero() bool   { return u.value == "" }
func (u Username) String() string { return u.value }
