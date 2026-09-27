# HTTP layer

`internal/api`

Everything shaped by HTTP: JSON bodies, routing, middleware, and turning
errors into responses. The use cases never see an `http.Request`.

The endpoints themselves, with their request and response bodies, are in the
[API reference](../api/reference.md).

## The use case adapter

Every endpoint backed by a use case decodes a request, runs the use case,
translates its error, logs, and writes a response. `adaptUseCase` does all of
that for any use case, given a decoder, an encoder and a success logger:

```go
adaptUseCase(useCaseAdapterParams[register.Input, register.Output]{
    UseCase:    cfg.Dependencies.Register,
    Decoder:    registerDecoder,
    Encoder:    registerEncoder,
    SuccessLog: registerSuccessLog,
})
```

A new endpoint is three small functions, not another handler with its own
copy of the error handling to get subtly wrong.

Before it decodes anything, the adapter answers `415` to a request whose
`Content-Type` is not `application/json`, so every endpoint it serves requires
that header ([API reference](../api/reference.md),
[why](../development/decisions/0051-post-endpoints-require-application-json.md)).
It then limits the body the decoder reads to 64 KiB (`requestBodyMaxBytes`,
with `http.MaxBytesReader`) and answers `413` when the decoder reaches the
limit, so every endpoint it serves has it
([API reference](../api/reference.md#request-bodies)). The wrap reads nothing
by itself, so the `415` still comes first.

## JWKS is a plain handler

It has no business logic: it reads the keys from a provider and reshapes
them. A use case for it would add a layer that only passes data along. The
keys it publishes come from the [keyring](tokens.md#signing-keys).

## Unknown routes

`NewRouter` wraps the `ServeMux` in `jsonRouteErrors`. When the mux has no
pattern for a request, the mux still answers it, and its plain-text `404` and
`405` are replaced with `ROUTE_NOT_FOUND` and `METHOD_NOT_ALLOWED` through
`writeJSON`, keeping the `Allow` header the mux computed
([Error model](../api/errors.md)). A catch-all `/` pattern would not do: it
would swallow the `405`s and their `Allow`. A response from a matched route,
and the mux's path-cleaning redirects, pass through untouched. The wrapper sits
inside `logging` and `recovery` and logs nothing of its own: `request finished`
carries the status.

## Middleware

Middleware wraps the router in this order: `logging`, then `recovery`.
`logging` takes the logger `NewRouter` was given, tags it with a request ID
and puts it into the request context, so every log line for the request
carries the same ID.
`recovery` catches a panic, logs it and answers `500` instead of dropping the
connection. It lets `http.ErrAbortHandler` through, since that panic is a
deliberate abort, not a bug. What each line carries is in
[Logging](logging.md).

## Error translation

`translateError` turns an error into a response:

| Error type | Result |
|---|---|
| `usecase.UseCaseError` | Logged with its `reason`; status from its `kind`, code from its `code` |
| `validation.ValidationError` | `422 VALIDATION_FAILED`, with an entry in `fields` for each failed field |
| anything else | Logged, and answered `500 INTERNAL_SERVER_ERROR` |

A [`UseCaseError`](usecases/README.md#usecaseerror) is logged at `info`, since
a rejected login is expected traffic rather than a defect. The line carries
the `code`, the `kind` and the `reason`, which is the only place a
deliberately generic code is told apart. Errors that carry no reason are
logged without the field.

Any other error is a bug, so it is logged as an error and its details are
kept from the client.

Every reason a generic code can hide is listed in
[Logging](logging.md#outcomes).

The shapes these responses take, and every code they can carry, are in the
[Error model](../api/errors.md).
