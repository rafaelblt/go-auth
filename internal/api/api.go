package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/infra"
)

type APIConfig struct {
	DBConnection string
	DevMode      bool
}

func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
	// Aplica de trás pra frente pra manter a ordem correta
	for i := len(middlewares) - 1; i >= 0; i-- {
		h = middlewares[i](h)
	}
	return h
}

func NewAPI(ctx context.Context, cfg APIConfig) (http.Handler, error) {
	pool, err := infra.NewPool(ctx, cfg.DBConnection)
	if err != nil {
		return nil, fmt.Errorf("new pool failed: %w", err)
	}

	deps, err := infra.NewDependencyContainer(ctx, infra.DependenciesConfig{
		DatabasePool: pool,
	})
	if err != nil {
		return nil, fmt.Errorf("dependencies creation failed: %w", err)
	}

	registerUseCase, err := deps.BuildRegister()
	if err != nil {
		return nil, fmt.Errorf("register use case creation failed: %w", err)
	}

	register := newRegisterHandler(registerUseCase)
	mux := http.NewServeMux()
	mux.HandleFunc("POST /auth/register", adaptHandler(register))

	middlewares := make([]func(http.Handler) http.Handler, 0)
	if !cfg.DevMode {
		middlewares = append(middlewares, adaptMiddleware(logging))
	}

	chain := chain(mux, middlewares...)

	return chain, nil
}
