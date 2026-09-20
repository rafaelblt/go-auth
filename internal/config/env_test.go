package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_Resolve_ReturnsNil_WhenEnvIsMissing(t *testing.T) {
	ev := env[int]{
		Key:    "GENERIC_VAR_MISSING",
		Parser: intEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	assert.Nil(t, value)
}

func TestEnv_Resolve_ReturnsPreset_WhenEnvMatches(t *testing.T) {
	presetKey := "prod"
	presetValue := 5432
	envKey := "GENERIC_VAR_PRESET"
	t.Setenv(envKey, presetKey)

	ev := env[int]{
		Key:     envKey,
		Presets: map[string]int{presetKey: presetValue},
		Parser:  intEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, presetValue, *value)
}

func TestEnv_Resolve_ParsesEnvValue(t *testing.T) {
	key := "GENERIC_VAR_PARSE"
	expected := 8080
	t.Setenv(key, fmt.Sprint(expected))

	ev := env[int]{
		Key:    key,
		Parser: intEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Equal(t, expected, *value)
}

func TestEnv_Resolve_ParsesEmptyEnvValue(t *testing.T) {
	key := "GENERIC_VAR_EMPTY"
	t.Setenv(key, "")

	ev := env[string]{
		Key:    key,
		Parser: stringEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	require.NotNil(t, value)
	assert.Empty(t, *value)
}

func TestEnv_Resolve_ReturnsErrorFromParser(t *testing.T) {
	expectedErr := errors.New("parse failed")
	key := "GENERIC_VAR_PARSE_ERROR"
	t.Setenv(key, "bad value")

	ev := env[int]{
		Key: key,
		Parser: func(value string) (int, error) {
			return 0, expectedErr
		},
	}

	value, err := ev.Resolve()

	assert.Nil(t, value)
	assert.ErrorIs(t, err, expectedErr)
}

func TestEnv_Resolve_PanicsWithParserNil(t *testing.T) {
	ev := env[string]{
		Key:    "GENERIC_VAR_WITHOUT_PARSER",
		Parser: nil,
	}

	assert.Panics(t, func() { ev.Resolve() })
}
