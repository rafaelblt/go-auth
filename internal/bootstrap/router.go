package bootstrap

import (
	"fmt"
	"net/http"

	"github.com/rafaelblt/go-auth/internal/api"
	"github.com/rafaelblt/go-auth/internal/config"
)

func newRouter(cfg config.Config, deps dependencies) (http.Handler, error) {
	apiCfg := api.Config{
		Dependencies: api.Dependencies{
			Logger:            deps.Logger,
			Clock:             deps.Infra.Clock,
			Register:          deps.UseCases.Register,
			Login:             deps.UseCases.Login,
			Refresh:           deps.UseCases.Refresh,
			Verify:            deps.UseCases.Verify,
			ChangePassword:    deps.UseCases.ChangePassword,
			PublicKeyProvider: deps.Infra.Ed25519Keyring,
		},
		RateLimiting: newRateLimiting(cfg, deps.Infra.RateLimiter),
	}

	router, err := api.NewRouter(apiCfg)
	if err != nil {
		return nil, fmt.Errorf("router creation failed: %w", err)
	}

	return router, nil
}
