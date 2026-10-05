package config

import (
	"errors"
	"fmt"

	"github.com/rafaelblt/go-auth/internal/validation"
)

var envKeyByField = map[string]string{
	fieldAddress:         envAddress.Key,
	fieldDatabaseURL:     envDatabaseURL.Key,
	fieldAutoMigrate:     envAutoMigrate.Key,
	fieldBcryptCost:      envBcryptCost.Key,
	fieldAccessTokenTTL:  envAccessTokenTTL.Key,
	fieldRefreshTokenTTL: envRefreshTokenTTL.Key,
	fieldLogFormat:       envLogFormat.Key,
	fieldRateLimit:       envRateLimit.Key,
	fieldTrustedProxies:  envTrustedProxies.Key,
}

// envLoad collects the failures of the conversion stage and remembers which
// field each one belongs to, so that the errors NewConfig reports for those
// same fields can be left out of the final report: their value never reached
// NewConfig, so its complaint would describe a consequence, not the cause.
type envLoad struct {
	errs   []error
	failed map[string]struct{}
}

func newEnvLoad() *envLoad {
	load := envLoad{
		errs:   make([]error, 0),
		failed: make(map[string]struct{}),
	}
	return &load
}

func resolveEnv[T any](load *envLoad, ev env[T], field string) *T {
	value, err := ev.Resolve()
	if err != nil {
		load.errs = append(load.errs, fmt.Errorf("'%s': %w", ev.Key, err))
		load.failed[field] = struct{}{}
	}
	return value
}

// merge appends the errors NewConfig reported to the conversion errors,
// rewriting each one in terms of the environment variable that carried it, so
// both stages are reported together instead of one round trip each.
func (load *envLoad) merge(err error) []error {
	if err == nil {
		return load.errs
	}

	var verr validation.ValidationError
	if !errors.As(err, &verr) {
		return append(load.errs, err)
	}

	errs := load.errs

	for _, fieldErr := range verr.Errors() {
		field := fieldErr.Field()
		if _, failed := load.failed[field]; failed {
			continue
		}
		errs = append(errs, fmt.Errorf("'%s': %s", envKeyOf(field), fieldErr.Issue()))
	}

	return errs
}

func envKeyOf(field string) string {
	key, ok := envKeyByField[field]
	if !ok {
		return field
	}
	return key
}
