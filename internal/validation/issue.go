package validation

import (
	"strings"
)

const (
	CodeTooLong  = "TOO_LONG"
	CodeTooShort = "TOO_SHORT"
)

type Issue struct {
	code   string
	params map[string]any
}

func (iss Issue) Error() string {
	return iss.code
}

type Issues []Issue

func (iss Issues) Error() string {
	msgs := make([]string, len(iss))

	for i, issue := range iss {
		msgs[i] = issue.code
	}

	return strings.Join(msgs, "; ")
}
