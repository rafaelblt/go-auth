package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/infra"
)

func NewRouter(ctx context.Context, conn string) (http.Handler, error) {
	pool, err := infra.NewPool(ctx, conn)
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

	register := NewRegisterHandler(registerUseCase)
	mux := http.NewServeMux()
    mux.Handle("POST /auth/register", register)

    return mux, nil
}
