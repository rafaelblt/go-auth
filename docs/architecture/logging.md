# Logging

`internal/bootstrap/run.go`, `internal/bootstrap/logger.go`,
`internal/api/middleware.go`, `internal/bootstrap/periodic.go`

Logs are the only record `go-auth` keeps of what it did: there is no audit
table and no metrics endpoint. This document lists every line the service can
write, what its fields mean, and how to work back from a response to the line
that explains it.

## The setup

One `*slog.Logger` serves the whole process. `bootstrap.Run` builds it from
[`LOG_FORMAT`](../configuration.md#log_format), as soon as the configuration is
known:

| | |
|---|---|
| Destination | standard output, one line per event |
| Format | `json` (default) or `text`, from `LOG_FORMAT` |
| Level | `info`. Handler options are `nil`, so `slog`'s default applies |
| Source positions | off |

The level is not configurable, and nothing in the service logs at `debug`, so
there is no verbosity to turn up. There is nothing to turn down either: the
quietest useful setting is the only setting.

That logger is then **passed by hand** down the composition root, like every
other dependency: `Run` gives it to `NewApp` in an `AppParams`, which keeps it
in its `dependencies` and hands it to `newRouter` and to the background tasks;
`NewRouter` rejects a `nil` one and gives it to the `logging` middleware. No
package reaches for `slog.Default()` to find it, so what the service writes is
decided by the configuration rather than by a global that some other package
may have replaced.

`Run` does still call `slog.SetDefault`, but nothing depends on it: it only
keeps stray `slog` calls, from a library or from code that has no logger in
hand, writing in the same format as everything else.

One line is written before all of this. **A configuration failure** is logged
by `slog.Default()` before the configured logger exists, so it goes to
**standard error** in Go's default format. If the service dies at startup and
standard output is empty, look at standard error.

`tests/e2e` passes no logger at all, so `NewApp` builds one from the test
config ([why that is allowed](startup.md#composition-root)), and its lines have
the same shape as the ones described here.

## Where logging happens

Only two places log: `internal/bootstrap` and `internal/api`.

The domain, the use cases, the ports and the infrastructure adapters do not
log at all. They return errors, and the HTTP layer logs the error once, where
the request ID and the outcome are both known. One failed request is one
error line, not one line per layer it fell through.

**Keep it that way when adding code.** A `*slog.Logger` passed into a use case
would make the same failure appear two or three times, with no way to tell a
repeat from a second request.

## Request correlation

`logging` is the outermost middleware. It gives every request a UUID, builds a
logger tagged with that ID and three facts about the request, and puts it in
the request context:

```json
{"time":"2026-09-22T12:29:28.334125303-03:00","level":"INFO","msg":"request received","request_id":"8f2c1e4a-7d33-4f8e-9a01-6b5c2d7e0f11","method":"POST","path":"/v1/auth/login","ip":"172.18.0.1:54122"}
```

| Field | Meaning |
|---|---|
| `request_id` | UUID v4, generated per request |
| `method` | HTTP method |
| `path` | Request path, without the query string |
| `ip` | `r.RemoteAddr`: the peer address as `host:port` |

Every line written while handling that request carries all four, because
handlers take their logger from the context with `loggerFrom(ctx)` rather than
from the default logger. `loggerFrom` falls back to `slog.Default()` when the
context carries no logger, which cannot happen inside a request: `logging`
wraps every route, and its context key is an unexported type, so no other
package can overwrite the value it stores. **Filtering by `request_id` gives the complete story
of one request**, and is the first move in any investigation.

Two things the fields do not tell you:

- **`ip` is the peer, not the client.** No proxy header is read for it, so
  behind the [reverse proxy the service expects](../limitations.md#run-it-behind-a-reverse-proxy)
  every request appears to come from the proxy. The client address is in the
  proxy's own log. Rate limiting does resolve the client address, reading
  `X-Forwarded-For` only when the peer is in
  [`TRUSTED_PROXIES`](../configuration.md#trusted_proxies), and its
  `rate limit exceeded` line carries it as `client_ip`.
- **`request_id` is generated here.** No inbound header is honoured, so it
  does not match an ID the proxy or the calling service assigned.

## Levels

The level is a claim about who should care, and it is applied consistently:

| Level | Means | Who acts |
|---|---|---|
| `INFO` | Something happened, including a request that was rejected | Nobody. Read when investigating |
| `WARN` | Handled, but needs attention: a catalog is incomplete, or a setting is deprecated or has no effect | The developer or the operator, eventually |
| `ERROR` | A defect or an operational failure | Whoever is on call |

A wrong password, a spent refresh token, a wrong content type, a rate-limited
request, an oversized body and a malformed body are all `INFO`.
They are the traffic an authentication service exists to reject, and paging on
them would page on every bot that finds the login endpoint.

The practical consequence, and the fastest triage rule in this codebase:

- **a `4xx` response never writes an `ERROR` line**;
- **a `5xx` response always writes one**, naming the defect.

So `level=ERROR` is a list of bugs, not a list of failed logins.

## Every line

### Request lifecycle

Written for every request, including `/.well-known/jwks.json`, which logs
nothing of its own.

| Message | Level | Extra fields | Written when |
|---|---|---|---|
| `request received` | `INFO` | — | Before the handler runs |
| `request finished` | `INFO` | `status`, `duration` | After it returns, including after a recovered panic |

`status` is what the handler wrote, defaulting to `200` if it wrote nothing.
`duration` is the whole middleware chain, measured around `ServeHTTP`. It is
**nanoseconds as a number in `json`**, and a string such as `62ms` in `text`.

A `request received` with no matching `request finished` means the panic was
`http.ErrAbortHandler`, which `recovery` deliberately re-panics. It escapes
both middlewares, so `net/http` handles it and logs it through the standard
`log` package, outside the format and the fields described here.

### Outcomes

One line per use case request, between the two lifecycle lines: one of these,
or one of the [defect](#defects) lines. A JWKS request produces neither.

| Message | Level | Extra fields | Response |
|---|---|---|---|
| `success register` | `INFO` | `user_id` | `201` |
| `success login` | `INFO` | `user_id`, `session_id` | `200` |
| `success refresh` | `INFO` | `user_id`, `session_id` | `200` |
| `unsupported media type error` | `INFO` | `content_type` | `415 unsupported_media_type` |
| `rate limit exceeded` | `INFO` | `client_ip`, `retry_after` | `429 too_many_requests` |
| `request body too large error` | `INFO` | — | `413 request_body_too_large` |
| `invalid json body error` | `INFO` | `error` | `400 invalid_json_body` |
| `validation error` | `INFO` | `pairs` | `422 validation_failed` |
| `use case error` | `INFO` | `code`, `kind`, `reason` | `401` or `409`, from `kind` |

`pairs` lists the failed fields as `"<field> <code>"`, one entry per
[`fields` entry](../api/errors.md#validation-errors), with the values
left out. It is an array in `json` and a bracketed string in `text`.

`content_type` is the `Content-Type` header as the client sent it, empty when
there was none.

`client_ip` is the [client address](http.md#client-address) the request was
counted against: an IPv4 address, or an IPv6 /64. `retry_after` is the wait
sent in `Retry-After`, in nanoseconds as a number in `json`, and a string such
as `2.5s` in `text`.

`reason` is the field to read, and the reason these lines exist. The API
answers several distinct failures with one deliberately generic code, so that
a client cannot tell them apart; the log is where they are told apart:

| Response code | `reason` | What actually failed |
|---|---|---|
| `invalid_credentials` | `malformed username` | Username could not be parsed |
| `invalid_credentials` | `malformed password` | Password could not be parsed |
| `invalid_credentials` | `user not found` | No account with that username |
| `invalid_credentials` | `password not found` | The account has no password row |
| `invalid_credentials` | `password mismatch` | Wrong password |
| `invalid_token` | `invalid token` | Refresh token unknown or malformed |
| `invalid_token` | `token expired` | Refresh token past its expiry |
| `invalid_token` | `token already used` | [Reuse detected](usecases/refresh.md#reuse-detection); the session was just revoked |
| `invalid_token` | `session revoked` | The session was already revoked |

Errors whose code already says everything, such as `username_already_exists`,
carry no reason, and the field is left out rather than logged empty.

`password not found` deserves attention: it means a user row exists without a
matching password row, which registration is written to make impossible.

### Defects

Every line here means something is wrong with the service. All but the first
answer `500`.

| Message | Level | Extra fields | Means | What to do |
|---|---|---|---|---|
| `error kind not found in message catalog, using fallback` | `WARN` | `kind` | A `kind` is missing from `kindMessageCatalog`. The client still gets the right status, with the text `An error occurred.` | Add the `kind` to the catalog |
| `error field not found in field catalog, using raw name` | `WARN` | `field` | A field is missing from `errorFieldCatalog`. The `422` body names it with its internal name, breaking the `snake_case` of every other field | Add the field to the catalog |
| `error kind not found in status catalog` | `ERROR` | `kind` | A `kind` is missing from `kindStatusCatalog`. A real answer was turned into a `500` | Add the `kind` to the catalog |
| `unexpected error for translation` | `ERROR` | `error` | An error reached the boundary that is neither a `UseCaseError` nor a `ValidationError`, usually from the database. `error` holds the text kept from the client | Read `error`. Usually the database: check it is reachable. Otherwise a missing translation, which is a bug |
| `rate limit check failed` | `ERROR` | `error` | The rate limiter returned an error, and the request was answered `500` without running. The in-memory limiter never fails | Read `error` |
| `panic recovered` | `ERROR` | `panic` | A handler panicked. `panic` is the recovered value | Always a bug. The `request_id` gives the request that triggered it |
| `json marshal failed` | `ERROR` | `error`, `body_type` | A response body could not be marshalled. The client gets `500` with an empty body | Always a bug, in the type named by `body_type` |

The first three are all the same kind of defect: something was added without
updating one of the three catalogs in `internal/api/errors.go`. The two
keyed by `ErrorKind` want a new `ErrorKind`; `errorFieldCatalog`, which maps a use
case's field name to the name that goes in a `422` body, wants a newly validated
field. None of the three fails in silence.

Severity follows what the client gets. A missing status mapping turns a real
answer into a `500`, so it is an `ERROR`. A missing message or field name still
answers correctly, with worse text, so both are `WARN`. Nothing reaches the field
path today: the only validated fields are register's two, and both are in the
catalog.

### Startup and shutdown

All written by `bootstrap.Run`, except where the table says otherwise.

| Message | Level | Extra fields | Means |
|---|---|---|---|
| `config load failed` | `ERROR` | `error` | Invalid configuration. **Standard error, default format.** Exit 1 |
| `building app...` | `INFO` | — | `NewApp` started |
| `app build failed` | `ERROR` | `error` | Wiring, migrations, the schema check or loading the signing keys failed. Exit 1 |
| `rate limiting on` | `INFO` | `rate_limit`, `trusted_proxies` | The [`RATE_LIMIT`](../configuration.md#rate_limit) level in force, and the proxies, from [`TRUSTED_PROXIES`](../configuration.md#trusted_proxies), whose `X-Forwarded-For` is read |
| `rate limiting is off; from v2 it cannot be turned off, only set to a level` | `WARN` | — | Rate limiting is off, the v1 default; v2 removes [`off`](../configuration.md#rate_limit) |
| `trusted proxies ignored while rate limiting is off` | `WARN` | `trusted_proxies` | [`TRUSTED_PROXIES`](../configuration.md#trusted_proxies) is set, but only rate limiting reads it, and it is off |
| `running app...` | `INFO` | — | The server is about to listen |
| `app run failed` | `ERROR` | `error` | The server failed, or graceful shutdown did. Exit 1 |
| `stopping app...` | `INFO` | — | Clean shutdown after `SIGINT` or `SIGTERM`. Exit 0 |

There is a gap between `running app...` and the first request line. Nothing is
logged when the listener is bound, so `running app...` is the readiness signal
available, and a port already in use surfaces as `app run failed`.

### Background tasks

All carry `task`, the task's name. There is one task,
`jwt_keyring_rotation`, which syncs the
[signing keys](tokens.md#signing-keys) every 10 minutes, with a 3 second
timeout per run. So `background task done` appears every 10 minutes on each
instance.

| Message | Level | Extra fields | Means |
|---|---|---|---|
| `background task started` | `INFO` | `interval` | The ticker started, at app start |
| `background task done` | `INFO` | `elapsed` | One run succeeded |
| `background task failed` | `ERROR` | `error`, `elapsed` | One run failed or timed out. The ticker keeps going |
| `background task stopped` | `INFO` | — | The app's context was cancelled |

A failed run is retried at the next tick, 10 minutes later. `elapsed` compared
with the task's timeout says whether the run failed or ran out of time: an
`elapsed` at or just over 3s means it timed out, and a shorter one means the
run itself returned an error.

A sync that keeps failing leaves the instance with the keys it last loaded. It
keeps signing, and switches to a key it already holds when that key's time
comes, but it does not see the keys that other instances add. If every
instance keeps failing, no key is added, and the current one signs past its 7
days, which is not dangerous.

A run deletes the retired keys only after it has published the others, so a
delete that fails, for instance because the database role cannot `DELETE`
from `signing_keys`, fails the run and nothing else: the retired keys stay in
the table until a delete succeeds. Startup leaves the deleting to the task, so
it does not fail for that.

## What is never logged

No line carries a credential or a token. This is a property of the call sites,
not of a filter, so it holds only as long as new lines respect it:

| Never logged | Where it would have been tempting |
|---|---|
| Passwords, plain or hashed | `validation error` logs field names and codes, never values |
| Access tokens | Success lines log no token; `json marshal failed` logs the body's **type**, not the body |
| Refresh token secrets | The same, and `use case error` describes the failure instead of the token |
| Usernames | Success lines identify the account by `user_id`, never by name |
| `DATABASE_URL` | Configuration errors name the variable, not its value |

Two supports for this. Secret-bearing value objects deliberately have no
`String` method, so `slog` cannot print them through `%v`
([why](conventions.md#value-rather-than-string)); and the only place a
configured value is echoed is a parse failure, which can only happen to a
number, a duration, a boolean or a trusted proxy entry, none of them secret.

Identifiers are the exception, and deliberately so: every success line carries
a UUID that names the account without describing it. `user_id` ties register,
login and refresh to one account, and `session_id` ties a login to the
refreshes that follow it. Failures carry neither, because the account behind a
rejected attempt is not always known and, when it is, saying so would put the
existence of an account in the log line.

## Reading the logs

With `LOG_FORMAT=json`, `jq` does the work. With `text`, `grep` does.

Everything about one request, in order:

```bash
jq -c 'select(.request_id == "8f2c1e4a-7d33-4f8e-9a01-6b5c2d7e0f11")' < logs.json
```

Every defect, which is the whole triage list:

```bash
jq -c 'select(.level == "ERROR")' < logs.json
```

Why a client is getting `401`, which the response body will not tell them:

```bash
jq -c 'select(.msg == "use case error") | {request_id, path, code, reason}' < logs.json
```

Requests slower than 500ms, remembering that `duration` is in nanoseconds:

```bash
jq -c 'select(.msg == "request finished" and .duration > 500000000)' < logs.json
```

Responses by status, to see whether a spike is rejections or defects:

```bash
jq -r 'select(.msg == "request finished") | .status' < logs.json | sort | uniq -c
```

Possible credential stuffing, as counts per peer address, which is only useful
when the service is not behind a proxy:

```bash
jq -r 'select(.reason == "password mismatch") | .ip' < logs.json | cut -d: -f1 | sort | uniq -c | sort -rn
```

Reuse detection firing, which revoked a session and logged someone out:

```bash
jq -c 'select(.reason == "token already used")' < logs.json
```

The life of one session, from the login that created it to its last refresh:

```bash
jq -c 'select(.session_id == "0b9f3f02-9a1f-4f7f-9a0e-4a1b6f2d9c77")' < logs.json
```

Everything one account did, across sessions:

```bash
jq -c 'select(.user_id == "2f1c8a5e-3b47-4f0a-9d6c-1e7b5a2c8d90")' < logs.json
```

## Adding a line

The conventions, in order of how often they are broken:

1. **Log at the HTTP boundary or in `bootstrap`**, never in a use case, the
   domain or an adapter.
2. **Take the logger from the context** with `loggerFrom(ctx)` inside
   `internal/api`, so the line carries the request ID. In `bootstrap`, take
   the `*slog.Logger` the function is given.
3. **The message is a short lowercase phrase**, and a constant: no
   interpolated values, so lines group. Everything variable is an attribute.
   The startup lines ending in `...` predate this and are the exception, not
   the pattern.
4. **Attribute keys are `snake_case`**, and reused as they are here:
   `error` for an error, `elapsed` and `duration` for times, `*_id` for
   identifiers.
5. **`ERROR` means a defect**, `WARN` means handled but needing attention (a
   degraded catalog, a deprecated or ineffective setting), and everything
   expected is `INFO`, however unwelcome it is to the client.
6. **Log no value that came from a request body**, and no token.

To assert on a line in a test, put a logger writing to a buffer into the
context under `loggerKey` and decode what comes out:

```go
buf := &bytes.Buffer{}
logger := slog.New(slog.NewJSONHandler(buf, nil))
ctx := context.WithValue(t.Context(), loggerKey, logger)
```

`internal/api/middleware_test.go` has the helpers that do this.

## What logs are not

- **Not an audit trail.** Lines are written to standard output and never
  stored by the service ([limitations](../limitations.md#out-of-scope)).
  Whatever collects them decides how long they live.
- **Not an interface.** Messages and fields change with the code. Alert on
  `level` and on the reasons above if you must, and expect to revisit them.
- **Not a metric.** There is no counter and no histogram; a rate means
  counting lines.
