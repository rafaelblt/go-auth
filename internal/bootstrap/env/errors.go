package env

import (
	"errors"
)

var ErrRequired = errors.New("the var is required but is missing")
