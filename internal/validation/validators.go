package validation

import (
	"fmt"
	"unicode/utf8"

	"github.com/rafaelblt/go-auth/internal/shared"
)

type Validator[T any] = func(value T) *Issue

func Validate[T any](value T, validators ...Validator[T]) Issues {
	var result Issues

	for _, validator := range validators {
		iss := validator(value)
		if iss != nil {
			result = append(result, *iss)
		}
	}

	return result
}

func MinLength(min int, unit LengthUnit) Validator[string] {
	length := lengthCounter(unit)

	return func(value string) *Issue {
		if length(value) < min {
			return shared.Ptr(IssueTooShort(min, unit))
		}
		return nil
	}
}

func MaxLength(max int, unit LengthUnit) Validator[string] {
	length := lengthCounter(unit)

	return func(value string) *Issue {
		if length(value) > max {
			return shared.Ptr(IssueTooLong(max, unit))
		}
		return nil
	}
}

func lengthCounter(unit LengthUnit) func(value string) int {
	switch unit {
	case UnitCodePoint:
		return utf8.RuneCountInString
	case UnitByte:
		return byteLength
	default:
		panic(fmt.Sprintf("unexpected length unit: %s", unit))
	}
}

func byteLength(value string) int { return len(value) }

func AllowedChars(allowed shared.Set[rune]) Validator[string] {
	return func(value string) *Issue {
		for _, char := range value {
			if !allowed.Contains(char) {
				return shared.Ptr(IssueInvalidChars())
			}
		}
		return nil
	}
}

func Required[T comparable]() Validator[T] {
	var zero T

	return func(value T) *Issue {
		if value == zero {
			return shared.Ptr(IssueRequired())
		}
		return nil
	}
}

type Number interface {
	~int | ~int64 | ~float64
}

func Positive[T Number]() Validator[T] {
	return func(value T) *Issue {
		if value <= 0 {
			return shared.Ptr(IssueNotPositive())
		}
		return nil
	}
}
