# Configuration

All configuration comes from environment variables, read once at startup by
`internal/config`. There is no config file, and nothing changes while the
process runs: a new value takes effect on restart.

If variables are missing or malformed, the service reports **all** of the
problems at once and exits, rather than starting half-configured.

## Reference

| Variable | Required | Default | Format |
|---|---|---|---|
| `DATABASE_URL` | yes | | PostgreSQL connection URL |
| `ADDRESS` | yes | | `host:port` to listen on |
| `AUTO_MIGRATE` | no | `false` | `true` / `false` |
| `BCRYPT_COST` | no | `12` | integer, or a preset |
| `ACCESS_TOKEN_TTL` | no | `30m` | Go duration, or a preset |
| `REFRESH_TOKEN_TTL` | no | `168h` (7 days) | Go duration, or a preset |
| `LOG_FORMAT` | no | `json` | `json` / `text` |

### `DATABASE_URL`

A standard PostgreSQL URL, passed to `pgxpool.ParseConfig`.

```
postgres://user:password@host:5432/dbname?sslmode=disable
```

The connection pool sets the session timezone to UTC and reads `timestamptz`
values in UTC, so timestamps are the same whatever the locale of the server
or the client.

Everything else about the pool uses pgx defaults, including the maximum and
minimum number of connections and their lifetimes. Those defaults depend on
the host's CPU count, so pool size varies between machines, and cannot be
configured.

An empty or whitespace-only value is rejected. This check matters: given an
empty connection string, pgx does *not* fail, but falls back to the libpq
environment variables and connects to whatever database they name.

### `ADDRESS`

The address the HTTP server listens on. Use `0.0.0.0:8080` inside a
container, and `localhost:8080` for local development.

The service serves plain HTTP. For TLS, put it behind a reverse proxy.

### `AUTO_MIGRATE`

When `true`, pending migrations are applied at startup. When `false`, the
default, the schema is only checked. Either way, the service does not start
unless the schema version matches the latest migration in the binary; see
[Migrations](getting-started.md#migrations).

Only `true` and `false` are accepted, in any case. `1`, `yes` and `on` are
errors.

### `BCRYPT_COST`

The bcrypt work factor. A higher cost is slower, and harder to crack offline.
It must be between 4 and 31, bcrypt's own limits. `internal/infra/bcrypt`
checks this range, not `internal/config`, so a value outside it fails when the
app is built, after the configuration has loaded.

| Preset | Value |
|---|---|
| `FAST` | 10 |
| `NORMAL` | 12 |
| `STRONG` | 14 |

Cost 12 takes roughly 200–400 ms per hash on typical hardware, and every
login pays it. So does startup, once, to create the dummy hash that login
checks when an account does not exist. Lower the cost only for tests; the
end-to-end tests use 6.

Stored passwords keep the cost they were hashed with. Changing this value
once there are users lets login timing reveal which accounts existed before
the change; see
[Limitations](limitations.md#login-timing-after-a-bcrypt_cost-change).

### `ACCESS_TOKEN_TTL`

How long an access token stays valid.

| Preset | Value |
|---|---|
| `SHORT` | 10m |
| `NORMAL` | 30m |
| `LONG` | 1h |

**The accepted range is 1 minute to 24 hours.** `internal/infra/jwt` checks
it, not `internal/config`, so a value outside it fails when the app is built,
with `new access token service failed: expiration must be between 1m and 24h`.

An access token cannot be revoked, so this is also how long a revoked session
keeps working. Keep it short. See
[Revocation is not immediate](api/token-verification.md#revocation-is-not-immediate).

### `REFRESH_TOKEN_TTL`

How long a refresh token stays valid. Every refresh issues a new token with
this lifetime, so this is how long a session can go unused before the client
must log in again, not how long a session can last.

| Preset | Value |
|---|---|
| `SHORT` | 24h |
| `NORMAL` | 168h (7 days) |
| `LONG` | 720h (30 days) |

It has no upper bound. It only has to be greater than zero.

### `LOG_FORMAT`

How log lines are written to standard output: `json`, the default, with one
JSON object per line, or `text`, as `key=value` pairs, which is easier to
read in a terminal. The value is case-insensitive.

Every line logged while handling a request carries a `request_id`, along with
the method, the path and the client address. Tokens, passwords and hashes are
never logged.

An error while loading the configuration is logged before the format is
known, so it goes to standard error in Go's default log format.

Every line the service can write, and how to read them, is in
[Logging](architecture/logging.md).

## Duration format

Durations are parsed with Go's `time.ParseDuration`, which accepts `ns`, `us`,
`ms`, `s`, `m` and `h`.

**There is no day unit.** `REFRESH_TOKEN_TTL=7d` stops the service from
starting. Write `168h` instead. Combined values such as `1h30m` work.

## Presets

Some variables accept a name instead of a value, for settings whose numbers
mean little without specialist knowledge:

```bash
BCRYPT_COST=STRONG
ACCESS_TOKEN_TTL=SHORT
```

A preset is looked up before the value is parsed, so its name does not have
to be a valid duration or number.

**Presets are uppercase and case-sensitive.** `ACCESS_TOKEN_TTL=short` is not
a preset: it is parsed as a duration, and fails with
`invalid duration value "short"`.

## Startup validation

Configuration is resolved in two stages, and both are reported together, so
one run lists every problem:

1. **Environment** (`LoadConfig`): every variable is looked up, matched
   against its presets, and parsed. A variable that is not set is not an error
   here: it resolves to "not given". Only values that fail to parse are
   collected at this stage.
2. **Values** (`NewConfig`): "not given" becomes the default for an optional
   variable, and the result is validated. Required values must not be empty,
   numbers and durations must be positive, and `LOG_FORMAT` must be one of the
   accepted values. This is where a missing `DATABASE_URL` or `ADDRESS` is
   caught.

`LoadConfig` ends by calling `NewConfig`, so a `Config` is built in one place
only, whether it comes from the environment or from code. It then merges the
errors of both stages, and rewrites each one in terms of the variable that
carried the value, so the report names the key you set rather than the field
it became:

```
invalid environment configuration: 'ACCESS_TOKEN_TTL': invalid duration value "7d": time: unknown unit "d" in duration "7d"
'DATABASE_URL': REQUIRED
```

A variable that failed to parse is reported once, by the first stage. Its
value never reached `NewConfig`, so a second complaint about the same variable
would describe a consequence rather than the cause.

`bootstrap.Run` logs the error, and the process exits with status 1. Nothing
starts half-configured.

## Values that are not configurable

| Setting | Value | Where |
|---|---|---|
| HTTP read header timeout | 5s | `internal/bootstrap/app.go` |
| HTTP read timeout | 15s | `internal/bootstrap/app.go` |
| HTTP write timeout | 15s | `internal/bootstrap/app.go` |
| HTTP idle timeout | 60s | `internal/bootstrap/app.go` |
| Graceful shutdown timeout | 10s | `internal/bootstrap/app.go` |
| Signing key rotation interval | 7 days | `internal/bootstrap/app.go` |
| Background task timeout, per run | 3s | `internal/bootstrap/app.go` |
| Schema version check timeout | 5s | `internal/bootstrap/schema.go` |
| Username length | 3–32 code points | `internal/domain/user/username.go` |
| Username characters | `a-z0-9._-` | `internal/domain/user/username.go` |
| Password minimum length | 8 code points | `internal/domain/password/plain.go` |
| Password maximum length | 72 bytes | `internal/domain/password/plain.go` |

The username and password rules are part of what the service is, not of how
it is deployed, so they are constants rather than settings. Their reasons are
in the [Domain model](architecture/domain/user.md#username).

## Configuring from code

Tests, and programs that build the app themselves, can skip the environment:

```go
cfg, err := config.NewConfig(config.ConfigParams{
    Address:         "localhost:8080",
    DatabaseURL:     "postgres://...",
    BcryptCost:      shared.Ptr(6),
    AccessTokenTTL:  shared.Ptr(30 * time.Minute),
    RefreshTokenTTL: shared.Ptr(24 * time.Hour),
})
```

Optional fields are pointers. For a number or a duration, zero is an invalid
setting, so a plain field could not tell "not given, use the default" from
"set to zero", which must be an error rather than quietly replaced by the
default. A `nil` pointer means not given. `AutoMigrate` is a plain `bool`,
because `false` is a valid setting.
