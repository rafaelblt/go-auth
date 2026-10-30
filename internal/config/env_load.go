package config

import "slices"

// envLoad collects the variables that failed to parse, so that LoadConfig can
// leave out what NewConfig then says about the same variables: their value
// never reached NewConfig, so its complaint would describe a consequence, not
// the cause.
type envLoad struct {
	problems []problem
	failed   map[string]struct{}
}

func newEnvLoad() *envLoad {
	load := envLoad{
		problems: make([]problem, 0),
		failed:   make(map[string]struct{}),
	}
	return &load
}

func resolveEnv[T any](load *envLoad, ev env[T]) *T {
	value, err := ev.Resolve()
	if err != nil {
		load.problems = append(load.problems, problem{key: ev.Key, err: err})
		load.failed[ev.Key] = struct{}{}
	}
	return value
}

// merge adds the problems NewConfig found to the parse failures, so both are
// reported together instead of one restart each.
func (load *envLoad) merge(problems []problem) []problem {
	merged := slices.Clone(load.problems)
	for _, p := range problems {
		if _, failed := load.failed[p.key]; !failed {
			merged = append(merged, p)
		}
	}
	return merged
}
