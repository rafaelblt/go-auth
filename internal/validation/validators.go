package validation

import (
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

func MinLength(min int) Validator[string] {
	return func(value string) *Issue {
		if utf8.RuneCountInString(value) < min {
			return shared.Ptr(IssueTooShort(min))
		}
		return nil
	}
}

func MaxLength(max int) Validator[string] {
	return func(value string) *Issue {
		if utf8.RuneCountInString(value) > max {
			return shared.Ptr(IssueTooLong(max))
		}
		return nil
	}
}

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
