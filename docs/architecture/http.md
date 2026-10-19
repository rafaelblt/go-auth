# HTTP

`internal/api`, `internal/infra/ratelimit`

Everything shaped by HTTP: routing, JSON bodies, middleware, rate limiting, and
turning errors into responses. The use cases never see an `http.Request`. The
contract the code serves is the [API reference](../api/reference.md).

| File | Holds |
|---|---|
| `router.go` | `Config`, `Dependencies`, `RateLimiting`, `NewRouter`: the routes, each endpoint's limit, the middleware |
| `adapters.go` | `adaptUseCase`, the body decoding (`decodeJSONBody`) and `writeJSON` |
| `endpoints.go` | per endpoint: request and response bodies, decoder, encoder, success log; the JWKS handler |
| `errors.go` | the error bodies, the three catalogs, `translateError` |
| `middleware.go` | `logging`, `recovery`, `jsonRouteErrors`, `loggerFrom` |
| `rate_limit.go` | `endpointRateLimit`, `clientAddress` |

## Use case adapter

`adaptUseCase` serves every endpoint backed by a use case, given a decoder, an
encoder and a success logger, so a new endpoint is three small functions in
`endpoints.go` rather than a handler with its own copy of the error handling
([Recipes](recipes.md#add-an-endpoint)). In order, it:

1. answers `415` when the `Content-Type` is not `application/json`, before
   anything is read ([Content type](#content-type));
2. when the endpoint has a rate limit, counts the request and answers `429`
   once the client address has used up its allowance
   ([Rate limiting](#rate-limiting));
3. limits the body to 64 KiB with `http.MaxBytesReader`, and decodes it with
   `decodeJSONBody`: `413` past the limit, `400` when it is not the expected
   JSON. What the decoder accepts and tolerates is part of the contract:
   [Request bodies](../api/reference.md#request-bodies);
4. runs the use case, then translates its error or encodes its output, logs
   the success, and writes the response with `writeJSON`.

Each step's log line is in [Logging](../logging.md#outcomes).

### Content type

Implements [Requests from browsers](../system.md#requests-from-browsers). The
check is the first thing the adapter does, and it answers without reading the
body. It sits in the adapter rather than in a middleware, which would have to
exempt the JWKS `GET` and would answer `415` before the mux answers `404`, or in
a wrapper on each route, which a new route could miss.

## Rate limiting

Implements [Rate limiting](../system.md#rate-limiting). The allowances of each
level are the table in [`RATE_LIMIT`](../configuration.md#rate_limit).

- `NewRouter` builds an `endpointRateLimit` for each endpoint `RateLimiting`
  gives a limit, and rejects an invalid one. The key is
  `<endpoint> <address>` (`login 203.0.113.9`), so each endpoint counts on its
  own.
- The levels' numbers are `rateLimitsByLevel` in
  `internal/bootstrap/rate_limit.go`, which builds `RateLimiting`, or none when
  `RATE_LIMIT` is `off`.
- The adapter asks right after the `415`, before the body is read. A limiter
  error is logged and answered `500`, and the use case does not run.
- Verify is given no limit on purpose ([why](../system.md#checking-an-access-token)).

**When you change it:** a new limited endpoint needs a field in `RateLimiting`
and `rateLimits`, a row in each level of `rateLimitsByLevel`, and a row in the
[`RATE_LIMIT`](../configuration.md#rate_limit) table.

### Client address

`clientAddress` implements [Which address counts](../system.md#which-address-counts):
`forwardedClient` walks `X-Forwarded-For` only when the peer is one of the
trusted prefixes, and `normalizeAddr` maps IPv4-mapped IPv6 to IPv4, drops the
zone and reduces IPv6 to its /64. A `RemoteAddr` that is not `ip:port` is used
as it is, and no header is read.

### The limiter

`internal/infra/ratelimit` (`in_memory.go`) implements `port.RateLimiter` with
`InMemory`, its only implementation. Per key it keeps one time, the moment the
key has its whole allowance back, which is how a limit of N per period allows N
requests at once, then one more every period/N; a refused request is not
counted. A sweep, at most once a minute, drops the keys past that time. The
counts live in the process
([Limitations](../system.md#rate-limiting-is-per-address-and-per-process)); a
store shared between instances would implement the same port.

## JWKS

`/.well-known/jwks.json` is `jwksHandler`, a plain handler: it reads the keys
from a `port.PublicKeyProvider` (the [keyring](tokens.md#signing-keys)) and
reshapes them, so a use case would only pass data along.

## Unknown routes

`NewRouter` wraps the `ServeMux` in `jsonRouteErrors`. When the mux has no
pattern for a request, its plain-text `404` and `405` are replaced with
`route_not_found` and `method_not_allowed`, keeping the `Allow` header the mux
computed. A catch-all `/` pattern would not do: it would swallow the `405`s.
Matched routes and the mux's path-cleaning redirects pass through untouched.

## Middleware

The router is wrapped in `logging`, then `recovery`, then `jsonRouteErrors`,
from the outside in ([A request, end to end](overview.md#a-request-end-to-end)).

- `logging` gives every request an ID, puts a logger tagged with it in the
  request context, and times the request with the clock it was given. Code in
  this package takes its logger with `loggerFrom(ctx)`
  ([Logging](conventions.md#logging)).
- `recovery` turns a panic into a logged `500`, except `http.ErrAbortHandler`, a
  deliberate abort, which it lets through. A panic in `logging` itself is outside
  it.

## Error translation

`translateError` turns a use case's error into a response:

| Error | Response |
|---|---|
| `usecase.UseCaseError` | status from its `kind`, code from its `code`; logged at `INFO` with its `reason` |
| `validation.ValidationError` | `422 validation_failed`, with an entry in `fields` per failed rule |
| anything else | `500 internal_server_error`; the error is logged at `ERROR`, and kept from the client |

Three catalogs in `errors.go` map the internal names to the response:
`kindStatusCatalog` and `kindMessageCatalog` by `ErrorKind`, and
`errorFieldCatalog` from a use case's field name to the name in the JSON body.

**When you change it:** a new `ErrorKind` or a newly validated field needs its
catalog entries; a missing one is logged ([Defects](../logging.md#defects)). The
response shapes and codes are the [Error model](../api/errors.md).
