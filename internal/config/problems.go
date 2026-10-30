package config

import (
	"fmt"
	"slices"
	"strings"
)

// problem is what is wrong with one setting, named by its environment
// variable whether the setting came from the environment or from code.
type problem struct {
	key string
	err error
}

// check records the first rule value breaks as the problem of the setting key
// names, so each setting has at most one.
func check[T any](problems *[]problem, key string, value T, rules ...func(T) error) {
	for _, rule := range rules {
		if err := rule(value); err != nil {
			*problems = append(*problems, problem{key: key, err: err})
			return
		}
	}
}

// configError is every problem found in the configuration. Its message is a
// report for whoever sets the variables: one line per setting, sorted by
// variable, each saying what the value must be.
type configError struct {
	problems []problem
}

func newConfigError(problems []problem) *configError {
	sorted := slices.Clone(problems)
	slices.SortStableFunc(sorted, func(a, b problem) int {
		return strings.Compare(a.key, b.key)
	})
	return &configError{problems: sorted}
}

func (e *configError) Error() string {
	width := 0
	for _, p := range e.problems {
		width = max(width, len(p.key))
	}

	settings := "settings"
	if len(e.problems) == 1 {
		settings = "setting"
	}

	var b strings.Builder
	fmt.Fprintf(&b, "invalid configuration, %d %s to fix:", len(e.problems), settings)
	for _, p := range e.problems {
		fmt.Fprintf(&b, "\n  %-*s  %v", width, p.key, p.err)
	}
	return b.String()
}
