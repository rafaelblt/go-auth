# Startup

`cmd`, `internal/bootstrap`, `internal/config`

How the process starts, wires everything, runs, and stops, and how the
configuration is loaded. The variables themselves, for whoever runs the service,
are in [Configuration](../configuration.md).

## Composition root

`internal/bootstrap` is the only package that knows how everything fits
together. `cmd/api.go` cancels a context on `SIGINT` or `SIGTERM`, calls
`bootstrap.Run`, and turns its error into exit status 1.

| File | Holds |
|---|---|
| `run.go` | `Run`: configuration, logger, `NewApp`, `App.Run`, and the startup log lines |
| `app.go` | `App`, `NewApp`, `App.Run` (the HTTP server and its timeouts), `App.Close`, the background tasks |
| `infra.go` | the pool, repositories, hasher, rate limiter, keyring, signer and token service; the signing key windows |
| `schema.go` | `autoMigrate`, `verifySchema` |
| `usecases.go` | one `build…` per use case, the dummy hash, the session cleanup's batch sizes |
| `router.go`, `rate_limit.go` | the router, and the limits of each `RATE_LIMIT` level |
| `periodic.go` | the background task runner |
| `logger.go` | `newLogger`, from `LOG_FORMAT` |

`Run` loads the configuration, builds the [logger](../logging.md#setup) from it,
sets it as `slog`'s default, and hands both to `NewApp`, which:

1. builds the infrastructure: pool, repositories, hasher, rate limiter;
2. applies the migrations, when `AUTO_MIGRATE` is set;
3. checks that the schema version is the latest embedded migration, and fails
   if not ([Migrations](persistence.md#migrations));
4. builds the keyring, which loads the signing keys and adds the first one when
   there is none, then the signer and the token service (`buildSigning`). It
   comes after the check because it reads the database;
5. builds the use cases, and the dummy password hash login and change password
   share;
6. builds the router, with the limits of the `RATE_LIMIT` level, or none when it
   is `off`.

**Rules the code keeps**

- Every `New` rejects a `nil` dependency or an invalid value
  ([Config struct](conventions.md#config-struct-with-a-validating-constructor)),
  so a wiring mistake stops the service here rather than on the first request
  that needs it. A setting reaches it already checked with the same rule
  ([Configuration loading](#configuration-loading)), so a value refused here
  is a wiring bug, not a configuration problem.
- A failure after the pool exists closes the pool before returning, through a
  deferred cleanup guarded by the named error; on success, `App.Close` owns it.
- The logger is the one dependency that may be `nil`: `NewApp` then builds it
  from the configuration's `LOG_FORMAT`, which is all a caller could have built
  it from. `tests/e2e` does that. From there it is passed by hand to the router
  and the background tasks.
- `App.Run` starts the background tasks and the HTTP server, and blocks until
  the server fails or the context is cancelled. On cancellation it lets
  in-flight requests finish, for up to the graceful shutdown time.

The values set here and not configurable, the HTTP timeouts among them, are
listed in [Configuration](../configuration.md#values-that-are-not-configurable).

## Background tasks

`periodic.go` runs a function on a ticker, with a timeout for each run, until
the app's context is cancelled. The first run of each task comes one interval
after startup, and a failed run is logged and retried at the next tick
([Background tasks](../logging.md#background-tasks)). Each run is timed with
`time.Now`, not the injected clock ([why](conventions.md#clock)). A task's
`run` returns, with its error, the attributes its `background task done` line
adds, as key-value pairs: `runSessionCleanup` turns the cleanup's `Output` into
them, and `withoutAttrs` wraps a function that has none to add, such as
`Keyring.Sync`.

| Task | Runs | Every | Timeout |
|---|---|---|---|
| `jwt_keyring_rotation` | [`Keyring.Sync`](tokens.md#signing-keys) | 10 minutes | 3 seconds |
| `session_cleanup` | the [session cleanup](usecases.md#session-cleanup) | hour | 5 minutes |

The keyring's first sync finds the keys startup already loaded. The session
cleanup is not registered at all while `SESSION_RETENTION` is `off`, which
builds no use case for it. Each of its runs deletes at most 10 batches of 1000
sessions, `sessionCleanupMaxBatches` and `sessionCleanupBatchSize` in
`usecases.go`. The tasks are listed in `App.backgroundTasks`.

**When you change it:** an interval or a batch size here is the one place it is
stated; [Session cleanup](../sessions.md#session-cleanup) and
[Signing keys](../access-tokens.md#signing-keys) link to this table. A new
task: [Recipes](recipes.md#add-a-background-task).

## Configuration loading

`internal/config` reads the environment once, at startup, checks every setting
with the rules of the code it is passed to, and reports every problem it finds
in one error, before `NewApp` builds anything.

| File | Holds |
|---|---|
| `config.go` | `Config` and its accessors, `ConfigParams`, `NewConfig`, `LoadConfig`, the defaults, and the rules each setting is checked with |
| `env_vars.go` | one `env[T]` per variable: its key, its parser, its presets |
| `env_var.go`, `env_load.go` | looking a variable up, applying presets, collecting parse failures |
| `parsers.go` | the parsers |
| `rules.go` | the rules `config` owns: `required`, `hostPort`, `oneOf` |
| `problems.go` | `check`, and the error that reports the problems |
| `log_format.go`, `rate_limit_level.go`, `session_retention.go` | the types with a closed set of values, or `off` |

It works in two stages, whose problems are reported together:

1. **Environment** (`LoadConfig`): each variable is looked up, matched against
   its presets and parsed. A variable that is not set is not a problem here;
   one that fails to parse is.
2. **Values** (`NewConfig`): an optional setting that was not given takes its
   default, and each setting is checked with its rules, in order, up to the
   first one it breaks.

**Rules the code keeps**

- A rule that the code a value is passed to relies on is that code's exported
  `Check…`, which its `New` calls too
  ([Setting rules](conventions.md#setting-rules)): `config` calls it, and never
  restates it. Its own rules are about the variable itself: `required`, the
  closed sets (`oneOf`), and `hostPort`, which is what `net.Listen` parses
  before it resolves anything.
- Each problem is named by the variable of its setting, whether the settings
  came from the environment or from code, and a setting has at most one: the
  first rule it breaks, or, for a variable that failed to parse, that failure.
- The error lists the problems sorted by variable, one per line.
  `bootstrap.Run` writes it to standard error as it is, since a log line would
  quote its line breaks. What the operator sees:
  [Startup validation](../configuration.md#startup-validation).
- No message holds the value of `DATABASE_URL` or `SIGNING_KEY_ENCRYPTION_KEY`:
  `postgres.CheckConnString` leaves out the string pgx quotes, and the key's
  errors give a size.

**Configuring from code.** Tests, and programs that build the app themselves,
skip the environment and call `NewConfig`:

```go
cfg, err := config.NewConfig(config.ConfigParams{
    Address:         "localhost:8080",
    DatabaseURL:     "postgres://...",
    BcryptCost:      shared.Ptr(6),
    AccessTokenTTL:  shared.Ptr(30 * time.Minute),
})
```

Optional settings are pointers, `nil` meaning "use the default": for a number or
a duration, zero is an invalid setting, and must be reported rather than
replaced by the default. `TrustedProxies` takes parsed `netip.Prefix` values,
and `SigningKeyEncryptionKey` the raw 32 bytes, `nil` meaning none for both.
`SessionRetention` takes `config.SessionRetentionOff()` or
`config.SessionRetentionOf(d)`.

**When you change it:** [Recipes](recipes.md#add-a-configuration-variable).

**Why:** a `Config` exists only once `NewConfig` accepts it, so checking the
settings in `bootstrap` would take a second report after the first is fixed,
or hand `NewApp` a `Config` it may refuse; restating the rules in `config`
would let the two copies drift apart. `config` imports the components' packages
for their rules instead.
