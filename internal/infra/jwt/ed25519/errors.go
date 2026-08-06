package ed25519

import (
	"errors"

	"github.com/rafaelblt/go-auth/internal/infra/jwt"
)

var (
	ErrTokenExpired = jwt.ErrTokenExpired
	ErrTokenInvalid = jwt.ErrTokenInvalid
	errKidMissing   = errors.New("token id missing")
	errKidUnknown   = errors.New("token id unknown")
)
