package api

import (
	"context"
	"errors"
	"net/http"
	"slices"

	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
)

type Config struct {
	Dependencies Dependencies
	DevMode      bool
}

type Dependencies struct {
	Register registerUseCase
	Login    loginUseCase
	Refresh  refreshUseCase
}

func NewRouter(ctx context.Context, cfg Config) (http.Handler, error) {
	if cfg.Dependencies.Register == nil {
		return nil, errors.New("register nil")
	}
	if cfg.Dependencies.Login == nil {
		return nil, errors.New("login nil")
	}
	if cfg.Dependencies.Refresh == nil {
		return nil, errors.New("refresh nil")
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

	for _, middleware := range slices.Backward(middlewares) {
		handler = middleware(handler)
	}

	return handler
}
