package config

import (
	"errors"
)

var errEnvRequired = errors.New("the var is required but is missing")
