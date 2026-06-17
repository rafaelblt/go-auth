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

	middlewares := make([]func(http.Handler) http.Handler, 0)
	if !cfg.DevMode {
		middlewares = append(middlewares, adaptMiddleware(logging))
	}

	chain := chain(mux, middlewares...)

	return chain, nil
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Aplica de trás pra frente pra manter a ordem correta
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}
