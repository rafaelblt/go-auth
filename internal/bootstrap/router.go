package bootstrap

import (
	"context"
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/api"
)

func newRouter(ctx context.Context, deps dependencies) (http.Handler, error) {
	cfg := api.Config{Dependencies: api.Dependencies{
		Logger:            deps.Logger,
		Register:          deps.UseCases.Register,
		Login:             deps.UseCases.Login,
		Refresh:           deps.UseCases.Refresh,
		PublicKeyProvider: deps.Infra.Ed25519Keyring,
	}}

	router, err := api.NewRouter(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("router creation failed: %w", err)
	}

	return router, nil
}
