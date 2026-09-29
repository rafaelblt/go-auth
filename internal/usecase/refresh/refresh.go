// Package refresh exchanges a refresh token for a new pair, and carries reuse
// detection: a token presented a second time revokes its whole session.
//
// See docs/architecture/usecases/refresh.md.
package refresh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/domain/session"
	"github.com/rafaelblt/go-auth/internal/domain/user"
	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase"
)

type Refresh struct {
	sessions     port.SessionReader
	accessIssuer port.AccessTokenIssuer
	tokens       port.RefreshTokenReader
	uow          port.UnitOfWork
	clock        port.Clock
	refreshTTL   time.Duration
}

type Input struct {
	RefreshToken string
}

type Output struct {
	UserID       string
	SessionID    string
	AccessToken  usecase.AccessTokenDTO
	RefreshToken usecase.RefreshTokenDTO
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

var ErrTokenAlreadyUsed = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"token already used",
)

var ErrSessionRevoked = usecase.NewErrorWithReason(
	"invalid_token",
	usecase.ErrorKindUnauthorized,
	"session revoked",
)

type accessIssued = port.AccessTokenIssued

func (uc *Refresh) Execute(ctx context.Context, in Input) (Output, error) {
	now := uc.clock.Now()

	token, err := uc.resolveToken(ctx, in.RefreshToken)
	if err != nil {
		return Output{}, err
	}

	sess, err := uc.getSessionOfToken(ctx, token)
	if err != nil {
		return Output{}, err
	}

	if sess.IsRevoked() {
		return Output{}, ErrSessionRevoked
	}

	err = token.Use(now)
	switch {
	case errors.Is(err, session.ErrTokenAlreadyUsed):
		return Output{}, uc.handleTokenReuse(ctx, sess, now)
	case errors.Is(err, session.ErrTokenExpired):
		return Output{}, ErrTokenExpired
	case err != nil:
		return Output{}, fmt.Errorf("unexpected error from token use: %w", err)
	}

	output, err := uc.rotate(ctx, sess, token, now)
	if errors.Is(err, session.ErrTokenAlreadyUsed) {
		// A concurrent refresh spent the token between the read and the write.
		// Losing that race counts as reuse, exactly like a replay found on read.
		//
		// See docs/development/decisions/0050-refresh-token-use-is-settled-at-write.md.
		return Output{}, uc.handleTokenReuse(ctx, sess, now)
	}
	return output, err
}

func (uc *Refresh) resolveToken(ctx context.Context, raw string) (*session.RefreshToken, error) {
	secret, err := session.ParseRefreshTokenSecret(raw)
	if err != nil {
		return nil, ErrTokenInvalid
	}

	token, err := uc.tokens.FindByHash(ctx, secret.Hash())
	if err != nil {
		return nil, fmt.Errorf("find refresh token by hash failed: %w", err)
	}
	if token == nil {
		return nil, ErrTokenInvalid
	}
	return token, nil
}

func (uc *Refresh) getSessionOfToken(ctx context.Context, token *session.RefreshToken) (*session.Session, error) {
	sess, err := uc.sessions.FindByID(ctx, token.SessionID())
	if err != nil {
		return nil, fmt.Errorf("find session by id failed: %w", err)
	}
	if sess == nil {
		return nil, errors.New("session of refresh token not found")
	}
	return sess, nil
}

// handleTokenReuse revokes the session of a reused token and reports the reuse.
// It never returns nil.
func (uc *Refresh) handleTokenReuse(ctx context.Context, sess *session.Session, now time.Time) error {
	sess.Revoke(now)
	err := uc.uow.Do(ctx, func(deps port.UowDeps) error {
		if err := deps.SessionWriter.Update(ctx, sess); err != nil {
			return fmt.Errorf("session writer failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("session revocation failed: %w", err)
	}

	return ErrTokenAlreadyUsed
}

func (uc *Refresh) rotate(
	ctx context.Context,
	sess *session.Session,
	used *session.RefreshToken,
	now time.Time,
) (Output, error) {
	accessData, err := uc.issueAccessToken(sess.UserID())
	if err != nil {
		return Output{}, err
	}

	newToken, newSecret, err := uc.createToken(sess.ID(), used.ID(), now)
	if err != nil {
		return Output{}, err
	}

	err = uc.saveTokens(ctx, used, newToken)
	if err != nil {
		return Output{}, err
	}

	output := Output{
		UserID:       sess.UserID().String(),
		SessionID:    sess.ID().String(),
		AccessToken:  usecase.MapAccessTokenIssuedToDTO(accessData),
		RefreshToken: usecase.MapRefreshTokenToDTO(newToken, newSecret),
	}
	return output, nil
}

func (uc *Refresh) issueAccessToken(userID user.ID) (accessIssued, error) {
	accessToken, err := uc.accessIssuer.Issue(port.AccessTokenPayload{
		UserID: userID,
	})
	if err != nil {
		return accessIssued{}, fmt.Errorf("access token issuer failed: %w", err)
	}
	return accessToken, nil
}

func (uc *Refresh) createToken(
	sessionID session.SessionID,
	parentID session.RefreshTokenID,
	now time.Time,
) (*session.RefreshToken, session.RefreshTokenSecret, error) {
	token, secret, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: sessionID,
		ParentID:  &parentID,
		CreatedAt: now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		e := fmt.Errorf("new refresh token failed: %w", err)
		return nil, session.RefreshTokenSecret{}, e
	}
	return token, secret, nil
}

func (uc *Refresh) saveTokens(ctx context.Context, used, new *session.RefreshToken) error {
	err := uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.RefreshTokenWriter.MarkUsed(ctx, used)
		if err != nil {
			return fmt.Errorf("mark used refresh token failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Add(ctx, new)
		if err != nil {
			return fmt.Errorf("add new refresh token failed: %w", err)
		}

		return nil
	})
	if err != nil {
		return fmt.Errorf("refresh tokens saving failed: %w", err)
	}
	return nil
}
