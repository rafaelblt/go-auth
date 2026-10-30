package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnvLoad_Merge_ReturnsParseProblems_WhenConfigIsValid(t *testing.T) {
	load := newEnvLoad()
	parseFailure := problem{key: envBcryptCost.Key, err: errors.New("parse failed")}
	load.problems = append(load.problems, parseFailure)

	problems := load.merge(nil)

	assert.Equal(t, []problem{parseFailure}, problems)
}

func TestEnvLoad_Merge_SkipsVariablesThatFailedToParse(t *testing.T) {
	load := newEnvLoad()
	parseFailure := problem{key: envBcryptCost.Key, err: errors.New("must be a whole number")}
	load.problems = append(load.problems, parseFailure)
	load.failed[envBcryptCost.Key] = struct{}{}
	addressMissing := problem{key: envAddress.Key, err: errors.New("must be set")}

	problems := load.merge([]problem{
		{key: envBcryptCost.Key, err: errors.New("must be between 4 and 31")},
		addressMissing,
	})

	assert.Equal(t, []problem{parseFailure, addressMissing}, problems)
}
