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
	UsernameMinCodePoints = 3
	UsernameMaxCodePoints = 32
)

var usernameAllowedChars = shared.NewSetFrom(
	[]rune("abcdefghijklmnopqrstuvwxyz0123456789._-")...,
)

var usernameValidators = []validation.Validator[string]{
	validation.MinLength(UsernameMinCodePoints, validation.UnitCodePoint),
	validation.MaxLength(UsernameMaxCodePoints, validation.UnitCodePoint),
	validation.AllowedChars(usernameAllowedChars),
}

func NewUsername(value string) (Username, validation.Issues) {
	normalized := normalizeUsername(value)

	issues := validation.Validate(normalized, usernameValidators...)

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
