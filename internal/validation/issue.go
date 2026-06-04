package validation

import (
	"strings"
)

const (
	CodeTooLong  = "TOO_LONG"
	CodeTooShort = "TOO_SHORT"
)

type Issue interface {
	Code() string
	Details() map[string]any
}

type Issues []Issue

func (iss Issues) Error() string {
	msgs := make([]string, len(iss))

	for i, issue := range iss {
		msgs[i] = issue.Code()
	}

	return strings.Join(msgs, "; ")
}
