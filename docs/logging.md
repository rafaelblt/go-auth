# Logging

Logs are the only record `go-auth` keeps of what it did: there is no audit
table and no metrics endpoint. This document lists every line the service
writes itself, what its fields mean, and how to work back from a response to the line
that explains it. The rules for adding a line are in the
[code conventions](architecture/conventions.md#logging).

## Setup

One `*slog.Logger`, built from [`LOG_FORMAT`](configuration.md#log_format) as
soon as the configuration is known, writes every line:

| | |
|---|---|
| Destination | standard output, one line per event |
| Format | `json` (default) or `text` |
| Level | `info`, and not configurable: nothing logs at `debug` |

**A configuration failure is the exception.** It is logged before the
configured logger exists, so it goes to **standard error** in Go's default
format. If the service dies at startup and standard output is empty, look
there.

Lines written through Go's standard `log` package, by `net/http` (such as
`http: Accept error: …`, when the process runs out of file descriptors) or by a
library, go through the same handler, in the same format, at `INFO`, with no
field beyond `msg`. They are not listed below.

## Request correlation

The `logging` middleware gives every request a UUID, and every line written
while handling it carries the same four fields:

```json
{"time":"2026-09-22T12:29:28.334125303-03:00","level":"INFO","msg":"request received","request_id":"8f2c1e4a-7d33-4f8e-9a01-6b5c2d7e0f11","method":"POST","path":"/v1/auth/login","ip":"172.18.0.1:54122"}
```

| Field | Meaning |
|---|---|
| `request_id` | UUID v4, generated per request |
| `method` | HTTP method |
| `path` | Request path, without the query string |
| `ip` | `r.RemoteAddr`: the peer address, as `host:port` |

Filtering by `request_id` gives the complete story of one request. Two things
the fields do not say:

- **`ip` is the peer, not the client.** Behind a reverse proxy every request
  comes from the proxy. Rate limiting does resolve the client address, and its
  `rate limit exceeded` line carries it as `client_ip`.
- **`request_id` is generated here.** No inbound header is honoured, and the ID
  is not returned to the client: to find a client's request, search by time,
  path and status.

## Levels

| Level | Means | Who acts |
|---|---|---|
| `INFO` | Something happened, including a request that was rejected | Nobody. Read when investigating |
| `WARN` | Handled, but needs attention: a catalog is incomplete, or a setting is deprecated or has no effect | The developer or the operator, eventually |
| `ERROR` | A defect or an operational failure | Whoever is on call |

A wrong password, a spent refresh token, a wrong content type, a rate-limited
request and a malformed body are all `INFO`: they are the traffic an
authentication service exists to reject. So **a `4xx` never writes an `ERROR`
line, and a `5xx` always writes one**, naming the defect: `level=ERROR` is a
list of bugs, not of failed logins.

## Every line

### Request lifecycle

Written for every request.

| Message | Level | Extra fields | Written when |
|---|---|---|---|
| `request received` | `INFO` | — | Before the handler runs |
| `request finished` | `INFO` | `status`, `duration` | After it returns, including after a recovered panic |

`status` is what the handler wrote, `200` if nothing. `duration` is
**nanoseconds as a number in `json`**, and a string such as `62ms` in `text`.

A `request received` with no `request finished` means the panic was
`http.ErrAbortHandler`, which `recovery` lets through to `net/http`. `net/http`
drops it without a line, so those two facts are all the log shows of it.

### Outcomes

One line per request to a use case endpoint, between the two lifecycle lines:
one of these, or a [defect](#defects).

| Message | Level | Extra fields | Response |
|---|---|---|---|
| `success register` | `INFO` | `user_id` | `201` |
| `success login` | `INFO` | `user_id`, `session_id` | `200` |
| `success change password` | `INFO` | `user_id` | `200` |
| `success refresh` | `INFO` | `user_id`, `session_id` | `200` |
| `success logout` | `INFO` | `user_id`, `session_id`, `reason` | `204` |
| `success verify` | `INFO` | `user_id` | `200` |
| `success me` | `INFO` | `user_id` | `200` |
| `unsupported media type error` | `INFO` | `content_type` | `415` |
| `rate limit exceeded` | `INFO` | `client_ip`, `retry_after` | `429` |
| `request body too large error` | `INFO` | — | `413` |
| `invalid json body error` | `INFO` | `error` | `400` |
| `validation error` | `INFO` | `pairs` | `422` |
| `use case error` | `INFO` | `code`, `kind`, `reason` | `401` or `409` |

`content_type` is the header as sent, empty when there was none. `client_ip` is
the [client address](system.md#which-address-counts) the request was
counted against, an IPv4 address or an IPv6 /64, and `retry_after` is a
duration, formatted like `duration`. `pairs` lists the failed fields as
`"<field> <code>"`, without the values.

`reason` is the field to read. The API answers several failures with one
deliberately generic code, and the log is where they are told apart. Login and
change password write the `invalid_credentials` reasons alike; `path` says
which:

| Code | `reason` | What failed |
|---|---|---|
| `invalid_credentials` | `malformed username` | The username breaks the registration rules |
| `invalid_credentials` | `malformed password` | The password breaks the registration rules |
| `invalid_credentials` | `user not found` | No account with that username |
| `invalid_credentials` | `password not found` | The account has no password row, which registration should make impossible |
| `invalid_credentials` | `password mismatch` | Wrong password |
| `invalid_credentials` | `password changed` | The password was replaced after it was checked: on login, by a change that committed first; on change password, by another change |
| `invalid_token` | `invalid token` | Refresh and logout: the token is unknown or malformed, or the [session cleanup](system.md#session-cleanup) deleted its session. Verify and me: the access token fails one of the [checks](system.md#checking-an-access-token), other than expiry |
| `invalid_token` | `token expired` | The token is past its expiry |
| `invalid_token` | `token already used` | Refresh: [reuse detected](system.md#reuse-detection); the session was just revoked |
| `invalid_token` | `session revoked` | Refresh: the session was already revoked |
| `invalid_token` | `user not found` | Me: the access token is valid, but its user is no longer stored |

Once the session cleanup has deleted a session, its tokens are unknown: a token
that logged `token expired`, `token already used` or `session revoked` before
logs `invalid token` from then on.

A code that already says everything, such as `username_already_exists`, has
no reason, and the field is left out.

`success logout` has a `reason` in two cases, each the one refresh logs for the
same state, so one query finds both ([Reading the logs](#reading-the-logs)):

- `token already used`: the token had already been spent and its session was
  still active, which is [reuse](system.md#reuse-detection). The logout revoked
  the session, as it does whatever the token's state.
- `session revoked`: the session was already revoked, by an earlier logout,
  reuse detection or a password change, and the logout changed nothing.

Either way it answered `204`.

### Defects

Every line here means something is wrong with the service. The `ERROR` lines
answer `500`.

| Message | Level | Extra fields | Means |
|---|---|---|---|
| `error kind not found in message catalog, using fallback` | `WARN` | `kind` | A `kind` is missing from `kindMessageCatalog`; the client gets the right status, with `An error occurred.` |
| `error field not found in field catalog, using raw name` | `WARN` | `field` | A field is missing from `errorFieldCatalog`; the `422` names it by its internal name |
| `error kind not found in status catalog` | `ERROR` | `kind` | A `kind` is missing from `kindStatusCatalog`; a real answer became a `500` |
| `unexpected error for translation` | `ERROR` | `error` | An error that is not a use case or validation error reached the boundary: usually the database, otherwise a missing translation |
| `rate limit check failed` | `ERROR` | `error` | The rate limiter failed; the request was not run. The in-memory limiter never fails |
| `panic recovered` | `ERROR` | `panic` | A handler panicked: always a bug |
| `json marshal failed` | `ERROR` | `error`, `body_type` | A response body could not be marshalled; the client gets a `500` with an empty body |

The first three mean something was added without its entry in one of the
catalogs of `internal/api/errors.go` ([Error translation](architecture/http.md#error-translation)).

### Startup and shutdown

| Message | Level | Extra fields | Means |
|---|---|---|---|
| `config load failed` | `ERROR` | `error` | Invalid configuration. **Standard error, default format.** Exit 1 |
| `building app...` | `INFO` | — | `NewApp` started |
| `app build failed` | `ERROR` | `error` | Wiring, migrations, the schema check or the signing keys failed. Exit 1 |
| `rate limiting on` | `INFO` | `rate_limit`, `trusted_proxies` | The [`RATE_LIMIT`](configuration.md#rate_limit) level, and the [trusted proxies](configuration.md#trusted_proxies) |
| `rate limiting is off; from v2 it cannot be turned off, only set to a level` | `WARN` | — | [`RATE_LIMIT`](configuration.md#rate_limit) is `off`, which is deprecated |
| `trusted proxies ignored while rate limiting is off` | `WARN` | `trusted_proxies` | `TRUSTED_PROXIES` is set, but only rate limiting reads it |
| `signing keys are stored unencrypted; from v2 SIGNING_KEY_ENCRYPTION_KEY is required` | `WARN` | — | [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key) is not set, which is deprecated |
| `running app...` | `INFO` | — | The server is about to listen |
| `app run failed` | `ERROR` | `error` | The server failed, or graceful shutdown did. Exit 1 |
| `stopping app...` | `INFO` | — | Clean shutdown after `SIGINT` or `SIGTERM`. Exit 0 |

Nothing is logged when the listener is bound, so `running app...` is the
readiness signal there is, and a port already in use shows up as
`app run failed`.

### Background tasks

All carry `task`: `jwt_keyring_rotation`, which syncs the
[signing keys](system.md#signing-keys), and `session_cleanup`,
which deletes [dead sessions](system.md#session-cleanup) unless
`SESSION_RETENTION` is `off`. Each instance writes `background task done` for
each task once per interval
([Background tasks](architecture/startup.md#background-tasks)).

`interval` and `elapsed` are durations, formatted like `duration`.

| Message | Level | Extra fields | Means |
|---|---|---|---|
| `background task started` | `INFO` | `interval` | The ticker started, with the app |
| `background task done` | `INFO` | `elapsed`; for `session_cleanup`, `deleted` and `limit_reached` | One run succeeded |
| `background task failed` | `ERROR` | `error`, `elapsed` | One run failed or timed out; the next tick tries again |
| `background task stopped` | `INFO` | — | The app's context was cancelled |

An `elapsed` at or just over the task's timeout means the run timed out; a
shorter one, that it returned an error. What a failing sync leaves
behind: [Signing keys](architecture/tokens.md#signing-keys). A failing cleanup
leaves the sessions it did not reach to the next run: each batch it deleted
before the failure stays deleted.

`deleted` is how many sessions a cleanup run deleted. `limit_reached` is `true`
when the run stopped at its [limit](architecture/startup.md#background-tasks)
of batches rather than at a short one, so dead sessions may be left over for
the next run. That is expected while a backlog drains, such as the one a
deployment that ran without the cleanup starts with; when it keeps happening,
the deployment is [falling behind](system.md#the-session-cleanup-can-fall-behind).

## What is never logged

No line carries a credential, a token or a key. This holds because of the call
sites, not because of a filter:

| Never logged | Where it would have been tempting |
|---|---|
| Passwords, plain or hashed | `validation error` logs field names and codes, never values |
| Access tokens and refresh token secrets | Success lines log IDs; `json marshal failed` logs the body's type, not the body |
| Usernames | Lines identify the account by `user_id`, never by name |
| Signing key seeds | Key errors name a generation and a size |
| `SIGNING_KEY_ENCRYPTION_KEY` | Its errors give a byte offset or a size |
| The password in `DATABASE_URL` | Configuration errors name the variable, not its value. A connection error at startup comes from pgx, which can show the host, user and database, with the password masked |

Otherwise, a configured value is echoed only when it fails to parse, and only
for a number, a duration, a boolean or a trusted proxy entry. Success lines
carry `user_id`, and login, refresh and logout `session_id`, which tie an
account's requests together; failures carry neither, since naming the account
behind a rejected attempt would put its existence in the log.

## Reading the logs

With `LOG_FORMAT=json`, `jq` does the work.

```bash
# Everything about one request, in order
jq -c 'select(.request_id == "8f2c1e4a-7d33-4f8e-9a01-6b5c2d7e0f11")' < logs.json

# Every defect: the whole triage list
jq -c 'select(.level == "ERROR")' < logs.json

# Why clients are getting 401
jq -c 'select(.msg == "use case error") | {request_id, path, code, reason}' < logs.json

# Reuse detected, by a refresh or a logout, either of which revoked a session
jq -c 'select(.reason == "token already used")' < logs.json

# The life of one session, from its login to its last refresh or its logout
jq -c 'select(.session_id == "0b9f3f02-9a1f-4f7f-9a0e-4a1b6f2d9c77")' < logs.json
```

Logs are not an audit trail: the service writes them to standard output and
keeps nothing, so whatever collects them decides how long they live. Nor are
they an interface: messages and fields change with the code. There is no
metric; a rate means counting lines.
