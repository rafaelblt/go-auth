package domain

import (
	"errors"
	"fmt"
)

type Username struct {
	stringVO
}

var ErrUsernameTooLong = errors.New("the username value is too long")
var ErrUsernameTooShort = errors.New("the username value is too short")

var UsernameMinLen = 3
var UsernameMaxLen = 15

func NewUsername(value string) (Username, error) {
	vo, err := newStringVO(value)
	if err != nil {
		if errors.Is(err, errStringVOEmpty) {
			return Username{}, ErrUsernameTooShort
		} else {
			return Username{}, fmt.Errorf("unexpected stringVO error: %w", err)
		}
	}
	if len(vo.value) < UsernameMinLen {
		return Username{}, ErrUsernameTooShort
	}
	if len(vo.value) > UsernameMaxLen {
		return Username{}, ErrUsernameTooLong
	}
	return Username{vo}, nil
}

func ValidateUsername(value string) []error {
	normalized := normalizeUsername(value)

	var errs []error

	if len(normalized) < UsernameMinLen {
		errs = append(errs, ErrUsernameTooShort)
	}
	if len(normalized) > UsernameMaxLen {
		errs = append(errs, ErrUsernameTooLong)
	}

	return errs
}

func normalizeUsername(value string) string {
	return normalizeStringVO(value)
}
