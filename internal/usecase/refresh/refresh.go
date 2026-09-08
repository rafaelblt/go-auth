package refresh

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/session"
	"github.com/rafaelblt/go-auth/internal/usecase"
	"github.com/rafaelblt/go-auth/internal/user"
)

type Refresh struct {
	sessions         port.SessionReader
	accessIssuer     port.AccessTokenIssuer
	refreshResolver  port.RefreshTokenResolver
	refreshGenerator port.RefreshTokenGenerator
	uow              port.UnitOfWork
	clock            port.Clock
	refreshTTL       time.Duration
}

type Input struct {
	RefreshToken string
}

type Output struct {
	AccessToken  usecase.AccessTokenDTO
	RefreshToken usecase.RefreshTokenDTO
}

var ErrTokenInvalid = usecase.NewErrorWithReason(
	"INVALID_TOKEN",
	usecase.ErrorKindUnauthorized,
	"invalid token",
)

var ErrTokenExpired = usecase.NewErrorWithReason(
	"INVALID_TOKEN",
	usecase.ErrorKindUnauthorized,
	"token expired",
)

var ErrTokenAlreadyUsed = usecase.NewErrorWithReason(
	"INVALID_TOKEN",
	usecase.ErrorKindUnauthorized,
	"token already used",
)

var ErrSessionRevoked = usecase.NewErrorWithReason(
	"INVALID_TOKEN",
	usecase.ErrorKindUnauthorized,
	"session revoked",
)

type accessIssued = port.AccessTokenIssued
type refreshGenerated = port.RefreshTokenGenerated

func (uc *Refresh) Execute(ctx context.Context, in Input) (Output, error) {
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

	err = token.Use(uc.clock.Now())
	switch {
	case err == nil:
		return uc.completeRefresh(ctx, sess, token)
	case errors.Is(err, session.ErrTokenAlreadyUsed):
		return Output{}, uc.tokenAlreadyUsed(ctx, token.SessionID())
	case errors.Is(err, session.ErrTokenExpired):
		return Output{}, ErrTokenExpired
	default:
		return Output{}, fmt.Errorf("unexpected error from token use: %w", err)
	}
}

func (uc *Refresh) resolveToken(ctx context.Context, raw string) (*session.RefreshToken, error) {
	token, err := uc.refreshResolver.Resolve(ctx, raw)
	if err != nil {
		if errors.Is(err, session.ErrTokenInvalid) {
			return nil, ErrTokenInvalid
		}
		e := fmt.Errorf("unexpected error from refresh token resolver: %w", err)
		return nil, e
	}
	return token, nil
}

func (uc *Refresh) getSessionOfToken(ctx context.Context, token *session.RefreshToken) (*session.Session, error) {
	sess, err := uc.sessions.FindByID(ctx, token.SessionID())
	if err != nil {
		return nil, fmt.Errorf("find session by id failed: %w", err)
	}
	if sess == nil {
		return nil, errors.New("refresh token session id not exists")
	}
	return sess, nil
}

func (uc *Refresh) tokenAlreadyUsed(ctx context.Context, sessID session.SessionID) error {
	sess, err := uc.sessions.FindByID(ctx, sessID)
	if err != nil {
		return fmt.Errorf("find session by id failed: %w", err)
	}
	if sess.IsRevoked() {
		return nil
	}

	sess.Revoke(uc.clock.Now())
	err = uc.uow.Do(ctx, func(deps port.UowDeps) error {
		if err := deps.SessionWriter.Update(ctx, sess); err != nil {
			return fmt.Errorf("update session failed: %w", err)
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("uow do failed: %w", err)
	}

	return ErrTokenAlreadyUsed
}

func (uc *Refresh) completeRefresh(ctx context.Context, sess *session.Session, used *session.RefreshToken) (Output, error) {
	accessData, refreshData, err := uc.generateTokens(sess.UserID())
	if err != nil {
		return Output{}, err
	}

	newToken, err := uc.createToken(refreshData.Hash, sess.ID(), used.ID())
	if err != nil {
		return Output{}, err
	}

	err = uc.saveTokens(ctx, used, newToken)
	if err != nil {
		return Output{}, err
	}

	output := Output{
		AccessToken: usecase.AccessTokenDTO{
			Value:     accessData.Token.Value(),
			ExpiresAt: accessData.ExpiresAt,
		},
		RefreshToken: usecase.RefreshTokenDTO{
			Value:     refreshData.Raw,
			ExpiresAt: newToken.ExpiresAt(),
		},
	}
	return output, nil
}

func (uc *Refresh) generateTokens(userID user.ID) (accessIssued, refreshGenerated, error) {
	accessToken, err := uc.accessIssuer.Issue(port.AccessTokenPayload{
		UserID: userID,
	})
	if err != nil {
		e := fmt.Errorf("access token issuer failed: %w", err)
		return accessIssued{}, refreshGenerated{}, e
	}

	refreshToken, err := uc.refreshGenerator.Generate()
	if err != nil {
		e := fmt.Errorf("refresh token generator failed: %w", err)
		return accessIssued{}, refreshGenerated{}, e
	}

	return accessToken, refreshToken, nil
}

func (uc *Refresh) createToken(
	hash session.RefreshTokenHash,
	sessionID session.SessionID,
	parentID session.RefreshTokenID,
) (*session.RefreshToken, error) {
	now := uc.clock.Now()
	token, err := session.NewRefreshToken(session.RefreshTokenCreationParams{
		SessionID: sessionID,
		Hash:      hash,
		ParentID:  &parentID,
		CreatedAt: now,
		ExpiresAt: now.Add(uc.refreshTTL),
	})
	if err != nil {
		return nil, fmt.Errorf("new refresh token failed: %w", err)
	}
	return token, nil
}

func (uc *Refresh) saveTokens(ctx context.Context, used, new *session.RefreshToken) error {
	return uc.uow.Do(ctx, func(deps port.UowDeps) error {
		err := deps.RefreshTokenWriter.Update(ctx, used)
		if err != nil {
			return fmt.Errorf("update used refresh token failed: %w", err)
		}

		err = deps.RefreshTokenWriter.Add(ctx, new)
		if err != nil {
			return fmt.Errorf("add new refresh token failed: %w", err)
		}

		return nil
	})
}
