# Configuration

All configuration comes from environment variables, read once at startup by
`internal/config`: there is no config file, and a new value takes effect on
restart. Every missing or malformed variable is reported in one error, and the
service exits rather than start half-configured
([Startup validation](#startup-validation)).

[`.env.example`](../.env.example) lists every variable with its default, as a
template to copy into wherever the deployment sets the environment. It is not
a config file: the service reads no `.env` file, and `docker-compose.dev.yml`
sets the development stack's variables itself.

| Variable | Required | Default | Format |
|---|---|---|---|
| `DATABASE_URL` | yes | | PostgreSQL connection URL |
| `ADDRESS` | yes | | `host:port` to listen on |
| `AUTO_MIGRATE` | no | `false` | `true` / `false` |
| `BCRYPT_COST` | no | `12` | integer, or a [preset](#durations-and-presets) |
| `ACCESS_TOKEN_TTL` | no | `30m` | [duration](#durations-and-presets), or a preset |
| `REFRESH_TOKEN_TTL` | no | `168h` (7 days) | duration, or a preset |
| `LOG_FORMAT` | no | `json` | `json` / `text` |
| `RATE_LIMIT` | no | `off` | `off` / `relaxed` / `normal` / `strict` |
| `TRUSTED_PROXIES` | no | none | comma-separated IP addresses and CIDR prefixes |
| `SIGNING_KEY_ENCRYPTION_KEY` | no | none | base64 of 32 bytes |
| `SESSION_RETENTION` | no | `720h` (30 days) | `off`, or a duration of at least `1h` |

### `DATABASE_URL`

A PostgreSQL URL, passed to `pgxpool.ParseConfig`:

```
postgres://user:password@host:5432/dbname?sslmode=disable
```

An empty or blank value is rejected: given one, pgx would not fail, but fall
back to the libpq environment variables and connect to whatever database they
name. The pool uses pgx's defaults for everything else, including its size,
which depends on the host's CPU count and cannot be configured.

### `ADDRESS`

The address the HTTP server listens on: `0.0.0.0:8080` inside a container,
`localhost:8080` for local development. The service serves plain HTTP; for TLS,
put it [behind a reverse proxy](limitations.md#run-it-behind-a-reverse-proxy).

### `AUTO_MIGRATE`

When `true`, pending migrations are applied at startup. Either way, the service
starts only when the schema version is the latest migration in the binary
([Migrations](getting-started.md#migrations)). Only `true` and `false` are
accepted, in any case.

### `BCRYPT_COST`

The bcrypt work factor: higher is slower, and harder to crack offline. It must
be between 4 and 31, bcrypt's own limits, which are checked when the app is
built rather than with the other variables.

| Preset | Value |
|---|---|
| `FAST` | 10 |
| `NORMAL` | 12 |
| `STRONG` | 14 |

Cost 12 takes roughly 200–400 ms per hash on typical hardware. A login pays it
once, a password change twice (checking the current password, hashing the new
one), and startup once, for the dummy hash a missing account is checked
against. Lower it only for tests; the end-to-end tests use 6.

Stored passwords keep the cost they were hashed with until the user changes
them. Changing it once there are users lets response times tell which accounts
existed before the change
([Limitations](limitations.md#login-timing-after-a-bcrypt_cost-change)).

### `ACCESS_TOKEN_TTL`

How long an access token stays valid, between 1 minute and 24 hours. The range
is checked when the app is built, as
`new access token service failed: expiration must be between 1m and 24h`.

| Preset | Value |
|---|---|
| `SHORT` | 10m |
| `NORMAL` | 30m |
| `LONG` | 1h |

An access token cannot be revoked, so this is also how long a revoked session
keeps working ([Limitations](limitations.md#access-tokens-cannot-be-revoked)).
Keep it short.

### `REFRESH_TOKEN_TTL`

How long a refresh token stays valid. Every refresh issues a new token with
this lifetime, so it is how long a session can go unused, not how long it can
last ([Limitations](limitations.md#sessions-have-no-maximum-age)). It has no
upper bound.

| Preset | Value |
|---|---|
| `SHORT` | 24h |
| `NORMAL` | 168h (7 days) |
| `LONG` | 720h (30 days) |

### `LOG_FORMAT`

`json`, one object per line, or `text`, `key=value` pairs that are easier to
read in a terminal; case-insensitive. Every line the service writes is in
[Logging](logging.md).

### `RATE_LIMIT`

How strictly register, login, change password, refresh and me are limited per
client address; case-insensitive. Verify, the JWKS, unknown paths and wrong
methods are never limited.

| Endpoint | `relaxed` | `normal` | `strict` |
|---|---|---|---|
| `POST /v1/auth/register` | 30 per hour | 10 per hour | 3 per hour |
| `POST /v1/auth/login` | 30 per minute | 10 per minute | 3 per minute |
| `POST /v1/auth/change-password` | 30 per hour | 10 per hour | 3 per hour |
| `POST /v1/auth/refresh` | 300 per minute | 100 per minute | 30 per minute |
| `POST /v1/auth/me` | 300 per minute | 100 per minute | 30 per minute |

"10 per minute" means up to 10 requests at once, then one more every 6 seconds:
the allowance refills steadily. Each endpoint keeps its own count, and every
`application/json` request counts, whatever its answer; one answered `415` never
does. Past the limit, the endpoint answers `429 too_many_requests` with a
`Retry-After` header in seconds, and the refused request is not counted.

The numbers: every login past the format checks pays a bcrypt comparison, so
`normal` holds an address to about 10 password guesses a minute, which a user
who mistypes a few times never notices. Registration and password changes are
rare and cost a bcrypt hash or comparison each, so they are counted per hour.
Refresh costs no bcrypt, and a client refreshes about once per
`ACCESS_TOKEN_TTL` per session, so `normal` covers about 3000 sessions behind
one address. Me costs no bcrypt either, only a token check and one indexed
read, so it gets refresh's numbers. The limits are constants in
`internal/bootstrap/rate_limit.go`.

The counts are per address and kept in memory
([Limitations](limitations.md#rate-limiting-is-per-address-and-per-process)),
and how the address is found is under [`TRUSTED_PROXIES`](#trusted_proxies).

**`off` is deprecated.** It is the v1 default, so upgrading changes nothing; v2
removes it, and `RATE_LIMIT` will only choose the level. While it is `off`,
startup logs a `WARN`.

### `TRUSTED_PROXIES`

The reverse proxies whose `X-Forwarded-For` header is believed:

```
TRUSTED_PROXIES=10.0.0.0/8, 192.0.2.10
```

By default there are none: the client address is the TCP peer, and no client
can choose the address it is counted as. When the peer is listed, the client is
the first address in `X-Forwarded-For`, read from the right, that is not listed
([how](architecture/http.md#client-address)). List every proxy hop between the
client and the service: the walk stops at the first one missing. An application
server that calls the service for its users, such as a backend relaying their
logins, is a proxy too: list it, and have it append each user's address to
`X-Forwarded-For`, or all of its users share one allowance.

Spaces around entries are ignored, and a bare address means that one address.
An IPv4-mapped IPv6 address (`::ffff:a.b.c.d`) counts as IPv4, as a bare
address or as a prefix of `/96` or longer. Only `X-Forwarded-For` is read, not
`Forwarded` or `X-Real-IP`, and only for rate limiting: the logs keep the peer.
While `RATE_LIMIT` is `off` it has no effect, and startup logs a `WARN`.

An entry that is neither an address nor a prefix, an empty entry, an address
with a zone, and an IPv4-mapped prefix shorter than `/96` stop the service from
starting:

```
invalid environment configuration: 'TRUSTED_PROXIES': invalid trusted proxy "proxy.local": ParseAddr("proxy.local"): unexpected character (at "proxy.local")
```

### `SIGNING_KEY_ENCRYPTION_KEY`

The AES-256 key that seals the seeds of the signing keys stored in the database
([Sealed seeds](architecture/tokens.md#sealed-seeds)): 32 random bytes, in
standard base64.

```bash
openssl rand -base64 32
```

Keep it out of the database and its backups. Once it is set, every signing key
added is sealed, and the keys stored before keep working until they are
deleted, about 9 days later: nothing needs migrating.

Every instance needs the same value, in the same rollout. Once sealed keys are
stored, an instance without it, or with another, does not start
(`app build failed`, with `seed sealed, but no encryption key set` or
`seed open failed: …`), and one already running without it fails every sync and
keeps signing with the keys it had. It cannot be changed or removed in place:
[replace the signing keys](#replacing-the-signing-keys) instead.

A value that is not standard base64, or not 32 bytes once decoded, stops the
service from starting, and the error names the variable, never the value:

```
invalid environment configuration: 'SIGNING_KEY_ENCRYPTION_KEY': decoded to 16 bytes, not 32
```

An empty value means not set. **Leaving it unset is deprecated.** It is the v1
default, so upgrading changes nothing; v2 requires it. Until then, startup logs
a `WARN`.

### `SESSION_RETENTION`

How long a session stays stored once it is dead, revoked or with its latest
refresh token expired, before the
[session cleanup](architecture/usecases.md#session-cleanup) deletes it, with its
refresh tokens. It counts from when the session died: it is not how long a
session can last, and has nothing to do with `REFRESH_TOKEN_TTL`.

`off`, in any case, keeps every session, and the cleanup never runs. Anything
else is a [duration](#durations-and-presets) of at least `1h`:

- **At least an hour**, because a refresh checks its session only at its start,
  and a rotation fails its writes, with a `500`, if the session is deleted
  before it ends. A rotation needs a live session, so the cleanup can delete
  one under it only if the refresh started more than the retention ago, which
  an hour rules out. Reuse detection needs no margin: it answers a session
  deleted under it as an unknown token
  ([Reuse detection](architecture/usecases.md#reuse-detection)).
- **`0` does not mean `off`.** It is a duration under the minimum, and stops
  the service from starting like any other, so that no value can be read both
  as "keep nothing" and as "delete nothing".

```
invalid environment configuration: 'SESSION_RETENTION': too_small(min=1h0m0s)
```

**The cleanup is on by default.** Before it existed, nothing was deleted: the
first runs after upgrading delete every session dead for more than 30 days, and
their refresh tokens, a little at a time
([Session cleanup](architecture/usecases.md#session-cleanup)). Set it to `off`
before upgrading to keep them.

## Replacing the signing keys

To change or remove `SIGNING_KEY_ENCRYPTION_KEY`, or to withdraw a signing key
that may have leaked:

1. stop every instance;
2. delete every row of `signing_keys`;
3. start the instances, with the new encryption key if that is the change.

The first instance to start adds a new key, which signs at once. Access tokens
signed before stop verifying here, and clients get new ones by refreshing.
Other verifiers keep the old keys until they fetch the JWKS again: until then
they still accept tokens signed with them, and accept the new ones only if they
refetch on an unknown `kid`
([caching](api/token-verification.md#fetching-and-caching-keys)).

## Durations and presets

Durations are parsed with Go's `time.ParseDuration`: `ns`, `us`, `ms`, `s`, `m`
and `h`, combined as in `1h30m`. **There is no day unit**: `7d` stops the
service from starting; write `168h`.

`BCRYPT_COST`, `ACCESS_TOKEN_TTL` and `REFRESH_TOKEN_TTL` also accept a preset
name, looked up before the value is parsed. **Presets are uppercase and
case-sensitive**: `ACCESS_TOKEN_TTL=short` is parsed as a duration, and fails.

## Startup validation

The configuration is resolved in two stages, reported together:

1. **Environment** (`LoadConfig`): each variable is looked up, matched against
   its presets and parsed. A variable that is not set is not an error here;
   one that fails to parse is.
2. **Values** (`NewConfig`): an optional variable that was not given takes its
   default, and the result is validated. Required values must not be empty,
   numbers and durations must be positive, `SESSION_RETENTION` at least `1h`
   unless it is `off`, and `LOG_FORMAT` and `RATE_LIMIT` must be one of their
   values.

The errors of both stages are merged into one, each named by the variable that
carried the value, and a variable that failed to parse is reported once, by the
first stage:

```
invalid environment configuration: 'ACCESS_TOKEN_TTL': invalid duration value "7d": time: unknown unit "d" in duration "7d"
'DATABASE_URL': required
```

It is logged to standard error, and the process exits with status 1. Three
checks come later, when the app is built, one failure at a time: the ranges of
`BCRYPT_COST` and `ACCESS_TOKEN_TTL`, and a `DATABASE_URL` made only of spaces.
They are logged as `app build failed` ([Logging](logging.md#startup-and-shutdown)).

## Values that are not configurable

| Setting | Value | Where |
|---|---|---|
| HTTP timeouts: read header, read, write, idle | 5s, 15s, 15s, 60s | `internal/bootstrap/app.go` |
| Graceful shutdown | 10s | `internal/bootstrap/app.go` |
| Schema version check | 5s | `internal/bootstrap/schema.go` |
| Request body limit | 64 KiB | [Request bodies](api/reference.md#request-bodies) |
| Signing key windows, sync interval and timeout | | [Signing keys](architecture/tokens.md#signing-keys) |
| Session cleanup interval, timeout and batches | | [Background tasks](architecture/startup.md#background-tasks) |
| Rate limits of each level | | [`RATE_LIMIT`](#rate_limit) |
| Username and password rules | | [Register](api/reference.md#post-v1authregister) |

## Configuring from code

Tests, and programs that build the app themselves, can skip the environment:

```go
cfg, err := config.NewConfig(config.ConfigParams{
    Address:         "localhost:8080",
    DatabaseURL:     "postgres://...",
    BcryptCost:      shared.Ptr(6),
    AccessTokenTTL:  shared.Ptr(30 * time.Minute),
})
```

Optional settings are pointers, `nil` meaning "use the default": for a number
or a duration, zero is an invalid setting, and must be reported rather than
replaced by the default. `TrustedProxies` takes parsed `netip.Prefix` values,
and `SigningKeyEncryptionKey` the raw 32 bytes, `nil` meaning none for both. A
key of another length fails when the app is built
(`new ed25519 keyring failed: encryption key size 16, not 32`).
`SessionRetention` takes `config.SessionRetentionOff()` or
`config.SessionRetentionOf(d)`.
