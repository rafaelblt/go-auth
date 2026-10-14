# Startup

`cmd`, `internal/bootstrap`

## Composition root

`internal/bootstrap` is the only package that knows how everything fits
together. `cmd/api.go` cancels a context on `SIGINT` or `SIGTERM`, calls
`bootstrap.Run`, and turns its error into exit status 1.

`Run` loads the configuration, builds the [logger](../logging.md#setup) from
it, and hands both to `NewApp`, which:

1. builds the infrastructure: pool, repositories, hasher, rate limiter
   (`infra.go`);
2. applies the migrations, when `AUTO_MIGRATE` is set (`schema.go`);
3. checks that the schema version is the latest embedded migration, and fails
   if not ([Migrations](persistence.md#migrations));
4. builds the keyring, which loads the signing keys and adds the first one when
   there is none, then the signer and the token service (`buildSigning`). It
   comes after the check because it reads the database;
5. builds the use cases, and the dummy password hash login and change password
   share (`usecases.go`);
6. builds the router, with the limits of the `RATE_LIMIT` level from
   `rate_limit.go`, or none when it is `off`.

Every `New` rejects a `nil` dependency or an invalid value
([Config struct](conventions.md#config-struct-with-a-validating-constructor)),
so a wiring mistake stops the service here rather than on the first request
that needs it. A failure after the pool exists closes the pool before
returning, through a deferred cleanup guarded by the named error; on success,
`App.Close` owns it.

The logger is the one dependency that may be `nil`: `NewApp` then builds it from
the configuration's `LOG_FORMAT`, which is all a caller could have built it
from. `tests/e2e` does that. From there the logger is passed by hand to the
router and the background tasks.

`App.Run` starts the background tasks and the HTTP server, and blocks until the
server fails or the context is cancelled. On cancellation it lets in-flight
requests finish, for up to 10 seconds.

## Background tasks

`periodic.go` runs a function on a ticker, with a timeout for each run, until
the app's context is cancelled. There is one task, `jwt_keyring_rotation`, which
runs [`Keyring.Sync`](tokens.md#signing-keys) every 10 minutes with a 3-second
timeout. The first run comes 10 minutes after startup, which has already loaded
the keys. A failed run is logged and retried at the next tick
([Background tasks](../logging.md#background-tasks)).
