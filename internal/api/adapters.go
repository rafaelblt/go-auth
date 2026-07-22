package api

import (
	"context"
	"net/http"
)

type middleware = func(http.Handler, http.ResponseWriter, *http.Request)

func adaptMiddleware(m middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m(next, w, r)
		})
	}
}

type response struct {
	StatusCode int
	Body       any
}

type useCase[Input, Output any] interface {
	Execute(context.Context, Input) (Output, error)
}

type decoder[Input any] = func(r *http.Request) (Input, error)
type encoder[Output any] = func(out Output) response
type successLog[Output any] func(context.Context, Output)

type useCaseAdapterParams[In, Out any] struct {
	UseCase    useCase[In, Out]
	Decoder    decoder[In]
	Encoder    encoder[Out]
	SuccessLog successLog[Out]
}

func adaptUseCase[In, Out any](p useCaseAdapterParams[In, Out]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := loggerFrom(ctx)

		input, err := p.Decoder(r)
		if err != nil {
			logger.Info("invalid json body error", "error", err)
			writeJSON(ctx, w, invalidJSONBodyError())
			return
		}

		output, err := p.UseCase.Execute(ctx, input)
		if err != nil {
			resp := translateError(ctx, err)
			writeJSON(ctx, w, resp)
			return
		}

		resp := p.Encoder(output)
		p.SuccessLog(ctx, output)
		writeJSON(ctx, w, resp)
	}
}
