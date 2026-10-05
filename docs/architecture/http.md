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
Next, when rate limiting is on, it counts the request against the endpoint's
allowance for the client address and answers `429` once that is used up, still
without reading the body ([Rate limiting](#rate-limiting)).
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
`405` are replaced with `route_not_found` and `method_not_allowed` through
`writeJSON`, keeping the `Allow` header the mux computed
([Error model](../api/errors.md)). A catch-all `/` pattern would not do: it
would swallow the `405`s and their `Allow`. A response from a matched route,
and the mux's path-cleaning redirects, pass through untouched. The wrapper sits
inside `logging` and `recovery` and logs nothing of its own: `request finished`
carries the status.

## Middleware

Middleware wraps the router in this order: `logging`, then `recovery`.
`logging` takes the logger and the clock `NewRouter` was given, tags the
logger with a request ID and puts it into the request context, so every log
line for the request carries the same ID. It times each request with that
clock.
`recovery` catches a panic, logs it and answers `500` instead of dropping the
connection. It lets `http.ErrAbortHandler` through, since that panic is a
deliberate abort, not a bug. What each line carries is in
[Logging](logging.md).

## Error translation

`translateError` turns an error into a response:

| Error type | Result |
|---|---|
| `usecase.UseCaseError` | Logged with its `reason`; status from its `kind`, code from its `code` |
| `validation.ValidationError` | `422 validation_failed`, with an entry in `fields` for each failed field |
| anything else | Logged, and answered `500 internal_server_error` |

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

## Rate limiting

Rate limiting is off unless [`RATE_LIMIT`](../configuration.md#rate_limit)
sets a level. It is checked in the use case adapter, right after the `415`, so
only the three `POST` endpoints are limited: JWKS, a `404` and a `405` never
are ([why](../development/decisions/0052-rate-limit-is-decided-before-the-body-is-read.md)).
An endpoint is limited only when `NewRouter` passes it a limit, so a new one
needs its limit there and a row in each level's table in
`internal/bootstrap/rate_limit.go`.

Every `application/json` request counts against the allowance, whatever its
answer. A request answered `415` is never counted: those are the requests a
browser sends cross-site without a preflight, and counting them would let any
web page spend its visitors' allowance. Each endpoint has its own allowance per
client address; the key is `<endpoint> <address>`, such as
`login 203.0.113.9`.

A refused request gets `429 too_many_requests` with `Retry-After`, and is not
counted. Nothing of it was read or run, so the same request can be sent again
after the wait. A limiter error is answered `500`, like any other failing
dependency.

### Client address

The client address starts as the TCP peer, `r.RemoteAddr`. `X-Forwarded-For`
is read only when the peer is in
[`TRUSTED_PROXIES`](../configuration.md#trusted_proxies): every header line is
joined, the entries are read from the right, and each trusted proxy is
skipped, so the first address that is not one is the client. When every entry
is trusted, the leftmost is the client. An empty entry, or one that is not a
bare address (`ip:port` included), stops the walk at the last good address.
Reading from the right is what keeps the address from being chosen by the
client: the leftmost entries are whatever the client sent, and only those a
trusted proxy appended can be believed.

Every address, the peer and each header entry, is normalised the same way: an
IPv4-mapped IPv6 address (`::ffff:a.b.c.d`) becomes IPv4, and a zone is
dropped. An IPv4 client counts as its address, and an IPv6 client as its /64,
since one host normally holds a whole /64 and could otherwise rotate through
it. A `RemoteAddr` that is not `ip:port` is used as it is, and no header is
read.

### The limiter

`port.RateLimiter` counts requests per key and says whether a request was
within its limit. `ratelimit.InMemory` is the only implementation. A limit of
N per period allows N requests at once, then one more every period/N; a
refused request is not counted.

Per key it keeps a single time: the moment at which the key has its whole
allowance back. A sweep, at most once a minute, drops the keys past that time,
so an idle key costs nothing. The counts live in the process, and a restart
forgets them. A store shared between processes would implement the same port.
