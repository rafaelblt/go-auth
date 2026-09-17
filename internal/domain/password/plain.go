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

func (p Plain) Value() string { return p.value }
func (p Plain) IsZero() bool  { return p.value == "" }
