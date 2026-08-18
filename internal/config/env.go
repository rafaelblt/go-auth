package config

import (
	"fmt"
	"os"
)

type env[T any] struct {
	Key      string
	Required bool
	Default  T
	Parser   func(string) (T, error)
	Presets  map[string]T
}

func (ev env[T]) Resolve() (T, error) {
	ev.validate()
	var zero T

	value, ok := ev.lookup()
	if !ok {
		if ev.Required {
			return zero, errEnvRequired
		} else {
			return ev.Default, nil
		}
	}

	preset, ok := ev.checkPreset(value)
	if ok {
		return preset, nil
	}

	parsed, err := ev.Parser(value)
	if err != nil {
		return zero, err
	}

	return parsed, nil
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
