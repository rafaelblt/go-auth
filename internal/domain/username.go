package domain

import (
	"errors"
	"fmt"
)

type Username struct {
	stringVO
}

var ErrUsernameEmpty = errors.New("the username value cannot be empty")

func NewUsername(value string) (Username, error) {
	vo, err := newStringVO(value)
	if err != nil {
		if errors.Is(err, errStringVOEmpty) {
			return Username{}, ErrUsernameEmpty
		} else {
			return Username{}, fmt.Errorf("unexpected stringVO error: %w", err)
		}
	}
	return Username{vo}, nil
}

func ValidateUsername(value string) []error {
	normalized := normalizeUsername(value)

	var errs []error

	if normalized == "" {
		errs = append(errs, ErrUsernameEmpty)
	}

	return errs
}

func normalizeUsername(value string) string {
	return normalizeStringVO(value)
}