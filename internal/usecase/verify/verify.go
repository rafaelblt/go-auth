// Package verify checks an access token for a service that would rather not
// verify it itself. It makes the checks a local verifier makes and nothing
// else: it reads no database, so the tokens of a revoked session still pass
// until they expire.
//
// See docs/architecture/usecases/verify.md.
package verify

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Verify struct {
	validator port.AccessTokenValidator
	clock     port.Clock
}

type Input struct {
	AccessToken string
}

type Output struct {
	UserID    string
	ExpiresAt time.Time
	ExpiresIn time.Duration
}

var ErrTokenInvalid = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"invalid token",
)

var ErrTokenExpired = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"token expired",
)

func (uc *Verify) Execute(_ context.Context, in Input) (Output, error) {
	now := uc.clock.Now()

	claims, err := uc.validator.Validate(in.AccessToken)
	switch {
	case errors.Is(err, session.ErrTokenInvalid):
		return Output{}, ErrTokenInvalid
	case errors.Is(err, session.ErrTokenExpired):
		return Output{}, ErrTokenExpired
	case err != nil:
		return Output{}, fmt.Errorf("access token validator failed: %w", err)
	}

	output := Output{
		UserID:    claims.UserID.String(),
		ExpiresAt: claims.ExpiresAt,
		ExpiresIn: claims.ExpiresAt.Sub(now).Truncate(time.Second),
	}
	return output, nil
}
