package config

import (
	"errors"
	"testing"

	"github.com/rafaelblt/go-auth/internal/validation"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestEnvLoad_Merge_ReturnsParseErrors_WhenConfigIsValid(t *testing.T) {
	load := newEnvLoad()
	parseErr := errors.New("parse failed")
	load.errs = append(load.errs, parseErr)

	errs := load.merge(nil)

	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], parseErr)
}

func TestEnvLoad_Merge_TranslatesFieldsToEnvKeys(t *testing.T) {
	load := newEnvLoad()
	verr := validation.NewValidationError(
		validation.NewFieldError(fieldBcryptCost, validation.IssueNotPositive()),
	)

	errs := load.merge(verr)

	require.Len(t, errs, 1)
	assert.Contains(t, errs[0].Error(), envBcryptCost.Key)
	assert.Contains(t, errs[0].Error(), validation.CodeNotPositive)
}

func TestEnvLoad_Merge_SkipsFieldsThatFailedToParse(t *testing.T) {
	load := newEnvLoad()
	parseErr := errors.New("invalid int value")
	load.errs = append(load.errs, parseErr)
	load.failed[fieldBcryptCost] = struct{}{}

	verr := validation.NewValidationError(
		validation.NewFieldError(fieldBcryptCost, validation.IssueRequired()),
		validation.NewFieldError(fieldAddress, validation.IssueRequired()),
	)

	errs := load.merge(verr)

	require.Len(t, errs, 2)
	assert.ErrorIs(t, errs[0], parseErr)
	assert.Contains(t, errs[1].Error(), envAddress.Key)
}

func TestEnvLoad_Merge_PassesThroughUnexpectedErrorShape(t *testing.T) {
	load := newEnvLoad()
	other := errors.New("something else")

	errs := load.merge(other)

	require.Len(t, errs, 1)
	assert.ErrorIs(t, errs[0], other)
}

func TestEnvKeyOf_FallsBackToTheFieldName(t *testing.T) {
	assert.Equal(t, "unmapped", envKeyOf("unmapped"))
}
