package config

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnv_Resolve_ReturnsDefault_WhenOptionalEnvIsMissing(t *testing.T) {
	fallback := "default fallback"
	ev := env[string]{
		Key:     "GENERIC_VAR_OPTIONAL_MISSING",
		Default: fallback,
		Parser:  stringEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	assert.Equal(t, fallback, value)
}

func TestEnv_Resolve_ReturnsZero_WhenOptionalEnvWithoutDefaultIsMissing(t *testing.T) {
	ev := env[int]{
		Key:    "GENERIC_VAR_OPTIONAL_WITHOUT_DEFAULT",
		Parser: intEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	assert.Equal(t, 0, value)
}

func TestEnv_Resolve_ReturnsError_WhenRequiredEnvIsMissing(t *testing.T) {
	key := "GENERIC_VAR_REQUIRED_MISSING"
	ev := env[string]{
		Key:      key,
		Required: true,
		Parser:   stringEnvParser,
	}

	value, err := ev.Resolve()

	assert.Zero(t, value)
	assert.ErrorIs(t, err, errEnvRequired)
}

func TestEnv_Resolve_ReturnsPreset_WhenEnvMatches(t *testing.T) {
	presetKey := "prod"
	presetValue := 5432
	envKey := "GENERIC_VAR_PRESET"
	t.Setenv(envKey, fmt.Sprint(presetKey))

	ev := env[int]{
		Key:     envKey,
		Presets: map[string]int{presetKey: presetValue},
		Parser:  intEnvParser,
	}

	value, err := ev.Resolve()

	require.NoError(t, err)
	assert.Equal(t, presetValue, value)
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
	assert.Equal(t, expected, value)
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

	assert.Zero(t, value)
	assert.ErrorIs(t, err, expectedErr)
}

func TestEnv_Resolve_PanicsWithParserNil(t *testing.T) {
	ev := env[string]{
		Key:    "GENERIC_VAR_WITHOUT_PARSER",
		Parser: nil,
	}

	assert.Panics(t, func() { ev.Resolve() })
}
