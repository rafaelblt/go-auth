package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net/http"
	"unicode/utf8"
)

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

// See docs/api/reference.md#request-bodies.
const requestBodyMaxBytes = 64 << 10

var errInvalidUTF8 = errors.New("request body is not valid UTF-8")

func hasJSONContentType(r *http.Request) bool {
	mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
	return err == nil && mediaType == "application/json"
}

// decodeJSONBody reads only the first JSON value, as before, and checks its
// bytes before encoding/json can turn invalid UTF-8 into U+FFFD. See
// docs/api/reference.md#request-bodies.
func decodeJSONBody(r *http.Request, body any) error {
	var raw json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&raw); err != nil {
		return err
	}
	if !utf8.Valid(raw) {
		return errInvalidUTF8
	}
	return json.Unmarshal(raw, body)
}

func writeJSON(ctx context.Context, w http.ResponseWriter, resp response) {
	bodyBuf, err := json.Marshal(resp.Body)
	if err != nil {
		loggerFrom(ctx).Error("json marshal failed", "error", err, "body_type", fmt.Sprintf("%T", resp.Body))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.StatusCode)
	w.Write(bodyBuf)
}
