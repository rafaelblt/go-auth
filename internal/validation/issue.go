package validation

import (
	"strings"
)

type Issue struct {
	code    string
	details map[string]any
}

func (iss Issue) Code() string {
	return iss.code
}

func (iss Issue) Details() map[string]any {
	return iss.details
}

type Issues []Issue

func (iss Issues) Error() string {
	msgs := make([]string, len(iss))

	for i, issue := range iss {
		msgs[i] = issue.Code()
	}

	return strings.Join(msgs, "; ")
}
