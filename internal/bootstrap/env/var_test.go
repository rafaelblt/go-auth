package env

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvVar_Resolve_ReturnsDefault_WhenOptionalEnvIsMissing(t *testing.T) {
	fallback := "default fallback"
	env := EnvVar[string]{
		Key:     "GENERIC_VAR_OPTIONAL_MISSING",
		Default: fallback,
		Parser:  StringParser,
	}

	value, err := env.Resolve()

	require.NoError(t, err)
	assert.Equal(t, fallback, value)
}

func TestEnvVar_Resolve_ReturnsZero_WhenOptionalEnvWithoutDefaultIsMissing(t *testing.T) {
	env := EnvVar[int]{
		Key:    "GENERIC_VAR_OPTIONAL_WITHOUT_DEFAULT",
		Parser: IntParser,
	}

	value, err := env.Resolve()

	require.NoError(t, err)
	assert.Equal(t, 0, value)
}

func TestEnvVar_Resolve_ReturnsError_WhenRequiredEnvIsMissing(t *testing.T) {
	key := "GENERIC_VAR_REQUIRED_MISSING"
	env := EnvVar[string]{
		Key:      key,
		Required: true,
		Parser:   StringParser,
	}

	value, err := env.Resolve()

	assert.Zero(t, value)
	assert.ErrorIs(t, err, ErrRequired)
}

func TestEnvVar_Resolve_ReturnsPreset_WhenEnvMatches(t *testing.T) {
	presetKey := "prod"
	presetValue := 5432
	envKey := "GENERIC_VAR_PRESET"
	t.Setenv(envKey, fmt.Sprint(presetKey))

	env := EnvVar[int]{
		Key:     envKey,
		Presets: map[string]int{presetKey: presetValue},
		Parser:  IntParser,
	}

	value, err := env.Resolve()

	require.NoError(t, err)
	assert.Equal(t, presetValue, value)
}

func TestEnvVar_Resolve_ParsesEnvValue(t *testing.T) {
	key := "GENERIC_VAR_PARSE"
	expected := 8080
	t.Setenv(key, fmt.Sprint(expected))

	gv := EnvVar[int]{
		Key:    key,
		Parser: IntParser,
	}

	value, err := gv.Resolve()

	require.NoError(t, err)
	assert.Equal(t, expected, value)
}

func TestEnvVar_Resolve_ReturnsErrorFromParser(t *testing.T) {
	expectedErr := errors.New("parse failed")
	key := "GENERIC_VAR_PARSE_ERROR"
	t.Setenv(key, "bad value")

	gv := EnvVar[int]{
		Key: key,
		Parser: func(value string) (int, error) {
			return 0, expectedErr
		},
	}

	value, err := gv.Resolve()

	assert.Zero(t, value)
	assert.ErrorIs(t, err, expectedErr)
}

func TestEnvVar_Resolve_PanicsWithParserNil(t *testing.T) {
	gv := EnvVar[string]{
		Key:    "GENERIC_VAR_WITHOUT_PARSER",
		Parser: nil,
	}

	assert.Panics(t, func() { gv.Resolve() })
}
