// Package validation holds the types for input rules: the Issue that names one
// broken rule, the validators that produce one, and the Accumulator that
// collects the failures of several fields. It serves two callers: the domain,
// whose failures the API answers with a 422, and config, whose failures become
// the startup error.
//
// See docs/architecture/domain/validation.md.
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

func (issues Issues) String() string {
	msgs := make([]string, len(issues))

	for i, issue := range issues {
		msgs[i] = issue.Code()
	}

	return strings.Join(msgs, "; ")
}

func (issues Issues) IsEmpty() bool {
	return len(issues) == 0
}
