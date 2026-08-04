package env

import (
	"fmt"
	"os"
)

type EnvVar[T any] struct {
	Key      string
	Required bool
	Default  T
	Parser   func(string) (T, error)
	Presets  map[string]T
}

func (ev EnvVar[T]) Resolve() (T, error) {
	ev.validate()
	var zero T

	value, ok := ev.lookup()
	if !ok {
		if ev.Required {
			return zero, ErrRequired
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

func (ev EnvVar[T]) validate() {
	if ev.Parser == nil {
		panic(fmt.Sprintf("the environment variable '%s' doesnt have a parser", ev.Key))
	}
}

func (ev EnvVar[T]) lookup() (string, bool) {
	return os.LookupEnv(ev.Key)
}

func (ev EnvVar[T]) checkPreset(value string) (T, bool) {
	if ev.Presets != nil {
		preset, ok := ev.Presets[value]
		if ok {
			return preset, true
		}
	}
	var zero T
	return zero, false
}
