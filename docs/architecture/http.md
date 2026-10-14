# HTTP

`internal/api`, `internal/infra/ratelimit`

Everything shaped by HTTP: routing, JSON bodies, middleware, rate limiting, and
turning errors into responses. The use cases never see an `http.Request`. The
endpoints themselves are in the [API reference](../api/reference.md).

## Use case adapter

`adaptUseCase` serves every endpoint backed by a use case, given a decoder, an
encoder and a success logger, so a new endpoint is three small functions rather
than a handler with its own copy of the error handling. In order, it:

1. answers `415` when the `Content-Type` is not `application/json`, before
   anything is read ([Content type](#content-type));
2. when the endpoint has a rate limit, counts the request and answers `429`
   once the client address has used up its allowance, still before reading the
   body ([Rate limiting](#rate-limiting));
3. limits the body to 64 KiB with `http.MaxBytesReader`, and decodes it: `413`
   past the limit, `400` when it is not the expected JSON
   ([Request bodies](../api/reference.md#request-bodies));
4. runs the use case, then translates its error or encodes its output, logs
   the success, and writes the response.

### Content type

A browser sends a cross-site `POST` without asking the site first when its
body is `text/plain`, a form or URL-encoded. For any other type,
`application/json` included, it first sends a CORS preflight, and `go-auth`
grants none: it sends no CORS headers at all. The decoders read any body as
JSON, so without this check a page on any site could send logins and
registrations through its visitors' browsers, each from its own address, past
any per-address rate limit, without needing to read the answers.

So the check is the first thing the adapter does, and it answers `415` without
reading the body. It sits in the adapter rather than in a middleware, which
would have to exempt the JWKS `GET` and would answer `415` before the mux
answers `404`, or in a wrapper on each route, which a new route could miss. A
reverse proxy that answers CORS preflights for these endpoints, for origins the
operator does not control, undoes it.

## Rate limiting

Off unless [`RATE_LIMIT`](../configuration.md#rate_limit) sets a level. Each
endpoint has its own allowance per client address, keyed as
`<endpoint> <address>` (`login 203.0.113.9`). Register, login, change password
and refresh have one, when `NewRouter` is given their limits; a new endpoint
needs its limit there and in each level's table in
`internal/bootstrap/rate_limit.go`.

The limit is decided in the adapter, right after the `415`:

- **After the `415`**, because the requests answered `415` are the ones a
  browser sends cross-site without a preflight: counting them would let any
  page spend its visitors' allowance.
- **Before the body is read**, so a refused request has read and run nothing.
  A `429` never spends a refresh token, and the same request can be sent again
  after `Retry-After`.
- **Every request past the `415` counts**, whatever its answer: a decision
  taken before the use case runs cannot know it.
- **A limiter error is a `500`**, and the use case does not run. Failing open
  would remove the limit without anyone noticing.

Verify has no limit, on purpose. It reads no database, and no number of tries
finds a signature that passes; a limit would only cost the services that
delegate verification, which call it on every request they serve, usually from
one address.

### Client address

The client address is the TCP peer, `r.RemoteAddr`. Only when the peer is in
[`TRUSTED_PROXIES`](../configuration.md#trusted_proxies) is `X-Forwarded-For`
read: its lines are joined, the entries are read from the right, and each
trusted proxy is skipped, so the client is the first address that is not one,
or the leftmost when all are. An empty entry, or one that is not a bare
address (`ip:port` included), stops the walk at the last good address. Reading
from the right is what keeps a client from choosing its address: the leftmost
entries are whatever it sent, and only those a trusted proxy appended can be
believed.

Every address is normalised alike: IPv4-mapped IPv6 becomes IPv4, and a zone is
dropped. An IPv6 client counts as its /64, since one host normally holds a
whole /64 and could otherwise rotate through it. A `RemoteAddr` that is not
`ip:port` is used as it is, and no header is read.

### The limiter

`port.RateLimiter` counts requests per key; `ratelimit.InMemory` is its only
implementation. A limit of N per period allows N requests at once, then one
more every period/N, and a refused request is not counted. Per key it keeps one
time, the moment the key has its whole allowance back, and a sweep, at most once
a minute, drops the keys past it. The counts live in the process: a restart
forgets them, and each instance counts on its own. A store shared between
instances would implement the same port.

## JWKS

`/.well-known/jwks.json` is a plain handler: it reads the keys from a
`port.PublicKeyProvider` (the [keyring](tokens.md#signing-keys)) and reshapes
them, so a use case would only pass data along.

## Unknown routes

`NewRouter` wraps the `ServeMux` in `jsonRouteErrors`. When the mux has no
pattern for a request, its plain-text `404` and `405` are replaced with
`route_not_found` and `method_not_allowed`, keeping the `Allow` header the mux
computed. A catch-all `/` pattern would not do: it would swallow the `405`s.
Matched routes and the mux's path-cleaning redirects pass through untouched.

## Middleware

The router is wrapped in `logging`, then `recovery`. `logging` gives every
request an ID, puts a logger tagged with it in the request context, and times
the request with the clock it was given ([Logging](../logging.md)). `recovery`
turns a panic into a logged `500`, except `http.ErrAbortHandler`, a deliberate
abort, which it lets through.

## Error translation

`translateError` turns a use case's error into a response:

| Error | Response |
|---|---|
| `usecase.UseCaseError` | status from its `kind`, code from its `code`; logged at `info` with its `reason` |
| `validation.ValidationError` | `422 validation_failed`, with an entry in `fields` per failed rule |
| anything else | `500 internal_server_error`; the error is logged, and kept from the client |

A rejected login is expected traffic, so a `UseCaseError` is `info`; anything
else is a bug, so it is `error`. Three catalogs in `internal/api/errors.go` map
the internal names to the response: `kindStatusCatalog` and
`kindMessageCatalog` by `ErrorKind`, and `errorFieldCatalog` from a use case's
field name to the name in the JSON body. A new `ErrorKind` or a newly validated
field needs its entries; a missing one is logged
([Defects](../logging.md#defects)). The shapes and codes of the responses are
in the [Error model](../api/errors.md).
