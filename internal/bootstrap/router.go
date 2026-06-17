package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/api"
)

func newRouter(ctx context.Context, uc usecases) (http.Handler, error) {
	cfg := api.Config{Dependencies: api.Dependencies{
		Register: uc.Register,
		Login:    uc.Login,
	}}

	router, err := api.NewRouter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("router creation failed: %w", err)
	}

	return router, nil
}
