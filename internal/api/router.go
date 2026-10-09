// Package api is the HTTP layer: routing, JSON request and response bodies,
// middleware, and turning errors into responses. The use cases never see an
// http.Request.
//
// See docs/architecture/http.md, and docs/api/reference.md for the endpoints
// themselves.
package api

import (
	"errors"
	"log/slog"
	"net/http"
	"net/netip"

	"github.com/rafaelblt/go-auth/internal/port"
	"github.com/rafaelblt/go-auth/internal/usecase/changepassword"
	"github.com/rafaelblt/go-auth/internal/usecase/login"
	"github.com/rafaelblt/go-auth/internal/usecase/refresh"
	"github.com/rafaelblt/go-auth/internal/usecase/register"
	"github.com/rafaelblt/go-auth/internal/usecase/verify"
)

type Config struct {
	Dependencies Dependencies
	// RateLimiting nil turns rate limiting off.
	RateLimiting *RateLimiting
}

type Dependencies struct {
	Logger            *slog.Logger
	Clock             port.Clock
	Register          useCase[register.Input, register.Output]
	Login             useCase[login.Input, login.Output]
	Refresh           useCase[refresh.Input, refresh.Output]
	Verify            useCase[verify.Input, verify.Output]
	ChangePassword    useCase[changepassword.Input, changepassword.Output]
	PublicKeyProvider port.PublicKeyProvider
}

type RateLimiting struct {
	Limiter        port.RateLimiter
	TrustedProxies []netip.Prefix
	Register       port.RateLimit
	Login          port.RateLimit
	Refresh        port.RateLimit
	ChangePassword port.RateLimit
}

func NewRouter(cfg Config) (http.Handler, error) {
	if cfg.Dependencies.Logger == nil {
		return nil, errors.New("logger nil")
	}
	if cfg.Dependencies.Clock == nil {
		return nil, errors.New("clock nil")
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
	if cfg.Dependencies.Verify == nil {
		return nil, errors.New("verify nil")
	}
	if cfg.Dependencies.ChangePassword == nil {
		return nil, errors.New("change password nil")
	}
	if cfg.Dependencies.PublicKeyProvider == nil {
		return nil, errors.New("public key provider nil")
	}

	var registerLimit, loginLimit, refreshLimit, changePasswordLimit *endpointRateLimit
	if rl := cfg.RateLimiting; rl != nil {
		if rl.Limiter == nil {
			return nil, errors.New("rate limiter nil")
		}
		if !validRateLimit(rl.Register) {
			return nil, errors.New("register rate limit invalid")
		}
		if !validRateLimit(rl.Login) {
			return nil, errors.New("login rate limit invalid")
		}
		if !validRateLimit(rl.Refresh) {
			return nil, errors.New("refresh rate limit invalid")
		}
		if !validRateLimit(rl.ChangePassword) {
			return nil, errors.New("change password rate limit invalid")
		}
		registerLimit = newEndpointRateLimit(*rl, "register", rl.Register)
		loginLimit = newEndpointRateLimit(*rl, "login", rl.Login)
		refreshLimit = newEndpointRateLimit(*rl, "refresh", rl.Refresh)
		changePasswordLimit = newEndpointRateLimit(*rl, "change-password", rl.ChangePassword)
	}

	register := adaptUseCase(useCaseAdapterParams[register.Input, register.Output]{
		UseCase:    cfg.Dependencies.Register,
		Decoder:    registerDecoder,
		Encoder:    registerEncoder,
		SuccessLog: registerSuccessLog,
		RateLimit:  registerLimit,
	})
	login := adaptUseCase(useCaseAdapterParams[login.Input, login.Output]{
		UseCase:    cfg.Dependencies.Login,
		Decoder:    loginDecoder,
		Encoder:    loginEncoder,
		SuccessLog: loginSuccessLog,
		RateLimit:  loginLimit,
	})
	refresh := adaptUseCase(useCaseAdapterParams[refresh.Input, refresh.Output]{
		UseCase:    cfg.Dependencies.Refresh,
		Decoder:    refreshDecoder,
		Encoder:    refreshEncoder,
		SuccessLog: refreshSuccessLog,
		RateLimit:  refreshLimit,
	})
	changePassword := adaptUseCase(useCaseAdapterParams[changepassword.Input, changepassword.Output]{
		UseCase:    cfg.Dependencies.ChangePassword,
		Decoder:    changePasswordDecoder,
		Encoder:    changePasswordEncoder,
		SuccessLog: changePasswordSuccessLog,
		RateLimit:  changePasswordLimit,
	})
	// Never rate limited. See docs/architecture/http.md#rate-limiting.
	verify := adaptUseCase(useCaseAdapterParams[verify.Input, verify.Output]{
		UseCase:    cfg.Dependencies.Verify,
		Decoder:    verifyDecoder,
		Encoder:    verifyEncoder,
		SuccessLog: verifySuccessLog,
	})

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/auth/register", register)
	mux.HandleFunc("POST /v1/auth/login", login)
	mux.HandleFunc("POST /v1/auth/refresh", refresh)
	mux.HandleFunc("POST /v1/auth/verify", verify)
	mux.HandleFunc("POST /v1/auth/change-password", changePassword)
	mux.Handle("GET /.well-known/jwks.json", &jwksHandler{cfg.Dependencies.PublicKeyProvider})

	handler := jsonRouteErrors(mux)
	handler = recovery(handler)
	handler = logging(cfg.Dependencies.Logger, cfg.Dependencies.Clock)(handler)

	return handler, nil
}
