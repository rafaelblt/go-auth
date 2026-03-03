package domain

import (
	"errors"
	"strings"
)

type stringVO struct {
	value string
}

var errStringVOEmpty = errors.New("string value object cannot be empty")

func newStringVO(value string) (stringVO, error) {
	normalized := normalizeStringVO(value)
	if normalized == "" {
		return stringVO{}, errStringVOEmpty
	}
	return stringVO{value: normalized}, nil
}

func normalizeStringVO(value string) string {
	return strings.TrimSpace(strings.ToLower(value))
}

func (s stringVO) String() string { return s.value }
func (s stringVO) IsZero() bool { return s.value == "" }
