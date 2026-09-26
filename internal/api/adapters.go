package api

import (
	"context"
	"errors"
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

		// See docs/development/decisions/0051-post-endpoints-require-application-json.md.
		if !hasJSONContentType(r) {
			logger.Info("unsupported media type error", "content_type", r.Header.Get("Content-Type"))
			writeJSON(ctx, w, unsupportedMediaTypeError())
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, requestBodyMaxBytes)
		input, err := p.Decoder(r)
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			logger.Info("request body too large error")
			writeJSON(ctx, w, requestBodyTooLargeError())
			return
		}
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
