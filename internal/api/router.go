// Package api is the HTTP layer: routing, JSON request and response bodies,
// middleware, and turning errors into responses. The use cases never see an
// http.Request.
//
// See docs/architecture/http.md, and docs/api/reference.md for the endpoints
// themselves.
package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"slices"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type Config struct {
	Dependencies Dependencies
}

type Dependencies struct {
	Logger            *slog.Logger
	Register          registerUseCase
	Login             loginUseCase
	Refresh           refreshUseCase
	PublicKeyProvider port.PublicKeyProvider
}

func NewRouter(ctx context.Context, cfg Config) (http.Handler, error) {
	if cfg.Dependencies.Logger == nil {
		return nil, errors.New("logger nil")
	}
	if cfg.Dependencies.Register == nil {
		return nil, errors.New("register nil")
	}
	if cfg.Dependencies.Login == nil {
		return nil, errors.New("login nil")
	}
	if cfg.Dependencies.Refresh == nil {
		return nil, errors.New("refresh nil")
	}
	if cfg.Dependencies.PublicKeyProvider == nil {
		return nil, errors.New("public key provider nil")
	}

	register := adaptUseCase(useCaseAdapterParams[register.Input, register.Output]{
		UseCase:    cfg.Dependencies.Register,
		Decoder:    registerDecoder,
		Encoder:    registerEncoder,
		SuccessLog: registerSuccessLog,
	})
	login := adaptUseCase(useCaseAdapterParams[login.Input, login.Output]{
		UseCase:    cfg.Dependencies.Login,
		Decoder:    loginDecoder,
		Encoder:    loginEncoder,
		SuccessLog: loginSuccessLog,
	})
	refresh := adaptUseCase(useCaseAdapterParams[refresh.Input, refresh.Output]{
		UseCase:    cfg.Dependencies.Refresh,
		Decoder:    refreshDecoder,
		Encoder:    refreshEncoder,
		SuccessLog: refreshSuccessLog,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", register)
	mux.HandleFunc("POST /v1/auth/login", login)
	mux.HandleFunc("POST /v1/auth/refresh", refresh)
	mux.Handle("GET /.well-known/jwks.json", &jwksHandler{cfg.Dependencies.PublicKeyProvider})

	chain := chainMiddlewares(mux, cfg.Dependencies.Logger)

	return chain, nil
}

func chainMiddlewares(handler http.Handler, logger *slog.Logger) http.Handler {
	type middleware func(http.Handler) http.Handler
	middlewares := []middleware{
		adaptMiddleware(logging(logger)),
		adaptMiddleware(recovery),
	}

	for _, middleware := range slices.Backward(middlewares) {
		handler = middleware(handler)
	}

	return handler
}
