package password

import (
	"github.com/rafaelblt/go-auth/internal/validation"
)

type Plain struct {
	value string
}

const (
	PlainMaxBytes      = 72
	PlainMinCodePoints = 8
)

var plainValidators = []validation.Validator[string]{
	validation.MinLength(PlainMinCodePoints, validation.UnitCodePoint),
	validation.MaxLength(PlainMaxBytes, validation.UnitByte),
}

func NewPlain(value string) (Plain, validation.Issues) {
	issues := validation.Validate(value, plainValidators...)

	if !issues.IsEmpty() {
		return Plain{}, issues
	}

	return Plain{value: value}, nil
}

// NewPlainReplacing reads a password meant to replace current: it follows the
// rules of NewPlain, and it is not current. Only current is compared, so a
// password used before it can be chosen again.
func NewPlainReplacing(value string, current Plain) (Plain, validation.Issues) {
	plain, issues := NewPlain(value)
	if !issues.IsEmpty() {
		return Plain{}, issues
	}
	if plain == current {
		return Plain{}, validation.Issues{validation.IssueSameAsCurrent()}
	}
	return plain, nil
}

func (p Plain) Value() string { return p.value }
func (p Plain) IsZero() bool  { return p.value == "" }
