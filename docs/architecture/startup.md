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
  that needs it.
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
`time.Now`, not the injected clock ([why](conventions.md#clock)).

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
stated; [How go-auth works](../system.md#processes) links to this table. A new
task: [Recipes](recipes.md#add-a-background-task).

## Configuration loading

`internal/config` reads the environment once, at startup, and reports every
problem it finds in one error.

| File | Holds |
|---|---|
| `config.go` | `Config` and its accessors, `ConfigParams`, `NewConfig`, `LoadConfig`, the defaults |
| `env_vars.go` | one `env[T]` per variable: its key, its parser, its presets |
| `env_var.go`, `env_load.go` | looking a variable up, applying presets, collecting parse errors |
| `parsers.go` | the parsers |
| `log_format.go`, `rate_limit_level.go`, `session_retention.go` | the types with a closed set of values or a minimum, and their validators |

It works in two stages, whose errors are merged:

1. **Environment** (`LoadConfig`): each variable is looked up, matched against
   its presets and parsed. A variable that is not set is not an error here; one
   that fails to parse is.
2. **Values** (`NewConfig`): an optional setting that was not given takes its
   default, and the result is validated with the
   [`validation`](domain.md#validation) types, one `Accumulator` for all of them.

Each error is named by the variable that carried the value, and a variable that
failed to parse is reported once, by the first stage. What the operator sees:
[Startup validation](../configuration.md#startup-validation).

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
and `SigningKeyEncryptionKey` the raw 32 bytes, `nil` meaning none for both; a
key of another length fails when the app is built. `SessionRetention` takes
`config.SessionRetentionOff()` or `config.SessionRetentionOf(d)`.

**When you change it:** [Recipes](recipes.md#add-a-configuration-variable).
