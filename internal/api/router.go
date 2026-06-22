package api

import (
	"context"
	"errors"
	"net/http"
)

type Config struct {
	Dependencies Dependencies
	DevMode      bool
}

type Dependencies struct {
	Register registerUseCase
	Login    loginUseCase
}

func NewRouter(ctx context.Context, cfg Config) (http.Handler, error) {
	if cfg.Dependencies.Register == nil {
		return nil, errors.New("register nil")
	}
	if cfg.Dependencies.Login == nil {
		return nil, errors.New("login nil")
	}

	register := newRegisterHandler(cfg.Dependencies.Register)
	login := newLoginHandler(cfg.Dependencies.Login)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", adaptHandler(register))
	mux.HandleFunc("POST /v1/auth/login", adaptHandler(login))

	chain := chainMiddlewares(cfg, mux)

	return chain, nil
}

func chainMiddlewares(cfg Config, handler http.Handler) http.Handler {
	type middleware func(http.Handler) http.Handler
	middlewares := make([]middleware, 0)

	if !cfg.DevMode {
		middlewares = append(middlewares, adaptMiddleware(logging))
	}
	middlewares = append(middlewares, adaptMiddleware(recovery))

	for i := len(middlewares) - 1; i >= 0; i-- {
		handler = middlewares[i](handler)
	}

	return handler
}
