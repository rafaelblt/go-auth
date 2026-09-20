package login

import (
	"context"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/password"
	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Input struct {
	Username string
	Password string
}

type Output struct {
	AccessToken  usecase.AccessTokenDTO
	RefreshToken usecase.RefreshTokenDTO
}

type Login struct {
	users        port.UserReader
	passwords    port.PasswordReader
	pwdChecker   port.PasswordChecker
	accessIssuer port.AccessTokenIssuer
	uow          port.UnitOfWork
	clock        port.Clock
	refreshTTL   time.Duration
	dummyHash    password.Hashed
}

var ErrUsernameMalformed = usecase.NewErrorWithReason(
	"INVALID_CREDENTIALS",
	usecase.ErrorKindUnauthorized,
	"malformed username",
)

var ErrPasswordMalformed = usecase.NewErrorWithReason(
	"INVALID_CREDENTIALS",
	usecase.ErrorKindUnauthorized,
	"malformed password",
)

var ErrUserNotFound = usecase.NewErrorWithReason(
	"INVALID_CREDENTIALS",
	usecase.ErrorKindUnauthorized,
	"user not found",
)

var ErrPasswordNotFound = usecase.NewErrorWithReason(
	"INVALID_CREDENTIALS",
	usecase.ErrorKindUnauthorized,
	"password not found",
)

var ErrPasswordMismatch = usecase.NewErrorWithReason(
	"INVALID_CREDENTIALS",
	usecase.ErrorKindUnauthorized,
	"password mismatch",
)

func (uc Login) Execute(ctx context.Context, input Input) (Output, error) {
	username, issues := user.NewUsername(input.Username)
	if !issues.IsEmpty() {
		return Output{}, ErrUsernameMalformed
	}
	plain, issues := password.NewPlain(input.Password)
	if !issues.IsEmpty() {
		return Output{}, ErrPasswordMalformed
	}

	usr, err := uc.users.FindByUsername(ctx, username)
	if err != nil {
		return Output{}, fmt.Errorf("find user by username failed: %w", err)
	}
	if usr == nil {
		return Output{}, uc.rejectWithDummyVerify(plain, ErrUserNotFound)
	}

	pwd, err := uc.passwords.FindByUserID(ctx, usr.ID())
	if err != nil {
		return Output{}, fmt.Errorf("find password by user id failed: %w", err)
	}
	if pwd == nil {
		return Output{}, uc.rejectWithDummyVerify(plain, ErrPasswordNotFound)
	}

	ok, err := uc.pwdChecker.Verify(plain, pwd.Hash())
	if err != nil {
		return Output{}, fmt.Errorf("password verification failed: %w", err)
	}
	if !ok {
		return Output{}, ErrPasswordMismatch
	}

	now := uc.clock.Now()

	sess, err := session.NewSession(session.SessionCreationParams{
		UserID:    usr.ID(),
		CreatedAt: now,
	})
	if err != nil {
		return Output{}, fmt.Errorf("session creation failed: %w", err)
	}

	issuedAccess, err := uc.accessIssuer.Issue(port.AccessTokenPayload{
		UserID: usr.ID(),
	})
	if err != nil {
		return Output{}, fmt.Errorf("access token issuer failed: %w", err)
	}

	refreshToken, refreshSecret, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: sess.ID(),
		ParentID:  nil,
		CreatedAt: now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		return Output{}, fmt.Errorf("refresh token creation failed: %w", err)
	}

	err = uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.SessionWriter.Add(ctx, sess)
		if err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Add(ctx, refreshToken)
		if err != nil {
			return fmt.Errorf("refresh token writer failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return Output{}, fmt.Errorf("session and refresh token saving failed: %w", err)
	}

	output := Output{
		AccessToken:  usecase.MapAccessTokenIssuedToDTO(issuedAccess),
		RefreshToken: usecase.MapRefreshTokenToDTO(refreshToken, refreshSecret),
	}
	return output, nil
}

// rejectWithDummyVerify spends the time of a real password check before
// returning rejection, so response time does not reveal whether the account
// exists.
func (uc Login) rejectWithDummyVerify(plain password.Plain, rejection error) error {
	_, err := uc.pwdChecker.Verify(plain, uc.dummyHash)
	if err != nil {
		return fmt.Errorf("dummy password verification failed: %w", err)
	}
	return rejection
}
