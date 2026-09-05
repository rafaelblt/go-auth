package credential

import (
	"github.com/rafaelblt/go-auth/internal/validation"
)

type PlainPassword struct {
	value string
}

const (
	PlainPasswordMaxBytes      = 72
	PlainPasswordMinCodePoints = 4
)

var plainPasswordValidators = []validation.Validator[string]{
	validation.MinLength(PlainPasswordMinCodePoints, validation.UnitCodePoint),
	validation.MaxLength(PlainPasswordMaxBytes, validation.UnitByte),
}

func NewPlainPassword(value string) (PlainPassword, validation.Issues) {
	issues := validation.Validate(value, plainPasswordValidators...)

	if !issues.IsEmpty() {
		return PlainPassword{}, issues
	}

	return PlainPassword{value: value}, nil
}

func (p PlainPassword) Value() string { return p.value }
func (p PlainPassword) IsZero() bool  { return p.value == "" }
