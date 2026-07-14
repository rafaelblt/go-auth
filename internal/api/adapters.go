package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

type response struct {
	StatusCode int
	Body       any
}

type handler interface {
	Handle(*http.Request) response
}

type middleware = func(http.Handler, http.ResponseWriter, *http.Request)

func adaptHandler(h handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		logger := loggerFrom(r.Context())

		resp := h.Handle(r)

		var buf []byte
		status := resp.StatusCode

		buf, err := json.Marshal(resp.Body)
		if err != nil {
			msg := "failed to encode http response body"
			logger.Error(msg, "error", err, "body_type", fmt.Sprintf("%T", resp.Body))
			buf, _ = json.Marshal(internalServerErrorBody)
			status = http.StatusInternalServerError
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		w.Write(buf)
	}
}

func adaptMiddleware(m middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			m(next, w, r)
		})
	}
}

type useCase[Input, Output any] interface {
	Execute(context.Context, Input) (Output, error)
}

type decoder[Input any] = func(r *http.Request) (Input, error)
type encoder[Output any] = func(w http.ResponseWriter, out Output) error

func adaptUseCase[In, Out any](
	uc useCase[In, Out],
	decoder decoder[In],
	encoder encoder[Out],
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		input, err := decoder(r)
		if err != nil {
			writeInvalidJSONBodyError(ctx, w)
			return
		}

		output, err := uc.Execute(ctx, input)
		if err != nil {
			translateErrorFromUseCase(ctx, w, err)
			return
		}

		err = encoder(w, output)
		if err != nil {
			writeInternalServerError(ctx, w)
		}
	}
}
