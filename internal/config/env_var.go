package config

import (
	"fmt"
	"os"
)

// env converts one environment variable from its raw string into T. It does
// not judge whether the value makes sense for the app: a variable that is not
// set resolves to nil, and NewConfig decides whether that means a default or
// a missing required value.
type env[T any] struct {
	Key     string
	Parser  func(string) (T, error)
	Presets map[string]T
}

func (ev env[T]) Resolve() (*T, error) {
	ev.validate()

	raw, ok := ev.lookup()
	if !ok {
		return nil, nil
	}

	preset, ok := ev.checkPreset(raw)
	if ok {
		return &preset, nil
	}

	parsed, err := ev.Parser(raw)
	if err != nil {
		return nil, err
	}

	return &parsed, nil
}

func (ev env[T]) validate() {
	if ev.Parser == nil {
		panic(fmt.Sprintf("the environment variable '%s' doesnt have a parser", ev.Key))
	}
}

func (ev env[T]) lookup() (string, bool) {
	return os.LookupEnv(ev.Key)
}

func (ev env[T]) checkPreset(value string) (T, bool) {
	if ev.Presets != nil {
		preset, ok := ev.Presets[value]
		if ok {
			return preset, true
		}
	}
	var zero T
	return zero, false
}
