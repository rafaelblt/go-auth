package config

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheck_RecordsOnlyTheFirstRuleBroken(t *testing.T) {
	var problems []problem
	first := errors.New("first")
	rules := []func(string) error{
		func(string) error { return nil },
		func(string) error { return first },
		func(string) error { return errors.New("second") },
	}

	check(&problems, "KEY", "value", rules...)

	assert.Equal(t, []problem{{key: "KEY", err: first}}, problems)
}

func TestCheck_RecordsNothing_WhenEveryRuleHolds(t *testing.T) {
	var problems []problem

	check(&problems, "KEY", "value", required)

	assert.Empty(t, problems)
}

func TestConfigError_ListsOneLinePerSetting_SortedAndAligned(t *testing.T) {
	err := newConfigError([]problem{
		{key: "RATE_LIMIT", err: errors.New("must be off, relaxed, normal or strict, not \"x\"")},
		{key: "ADDRESS", err: errors.New("must be set")},
	})

	assert.Equal(t, `invalid configuration, 2 settings to fix:
  ADDRESS     must be set
  RATE_LIMIT  must be off, relaxed, normal or strict, not "x"`, err.Error())
}

func TestConfigError_CountsASingleSetting(t *testing.T) {
	err := newConfigError([]problem{{key: "ADDRESS", err: errors.New("must be set")}})

	assert.Equal(t, "invalid configuration, 1 setting to fix:\n  ADDRESS  must be set", err.Error())
}
