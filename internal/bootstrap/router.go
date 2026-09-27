package bootstrap

import (
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/api"
)

func newRouter(deps dependencies) (http.Handler, error) {
	cfg := api.Config{Dependencies: api.Dependencies{
		Logger:            deps.Logger,
		Clock:             deps.Infra.Clock,
		Register:          deps.UseCases.Register,
		Login:             deps.UseCases.Login,
		Refresh:           deps.UseCases.Refresh,
		PublicKeyProvider: deps.Infra.Ed25519Keyring,
	}}

	router, err := api.NewRouter(cfg)
	if err != nil {
		return nil, fmt.Errorf("router creation failed: %w", err)
	}

	return router, nil
}
