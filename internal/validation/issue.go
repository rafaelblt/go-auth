package validation

import (
	"fmt"
	"maps"
	"slices"
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
	return maps.Clone(iss.details)
}

func (iss Issue) String() string {
	if len(iss.details) == 0 {
		return iss.code
	}

	pairs := make([]string, 0, len(iss.details))

	for _, key := range slices.Sorted(maps.Keys(iss.details)) {
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, iss.details[key]))
	}

	return iss.code + "(" + strings.Join(pairs, ", ") + ")"
}

type Issues []Issue

func (iss Issues) String() string {
	msgs := make([]string, len(iss))

	for i, issue := range iss {
		msgs[i] = issue.Code()
	}

	return strings.Join(msgs, "; ")
}

func (issues Issues) IsEmpty() bool {
	return len(issues) == 0
}
