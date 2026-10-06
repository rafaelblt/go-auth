# Startup

`cmd/api.go`, `internal/bootstrap`

## Composition root

`internal/bootstrap` is the only package that knows how everything fits
together, and `Run` is the whole of it: `cmd/api.go` installs the signal
handler and turns the returned error into an exit code, nothing else.

`Run` loads the configuration, builds the [logger](logging.md#the-setup) from
it, and hands both to `NewApp` in an `AppParams`, which runs, in order:

1. Build the infrastructure: pool, repositories, hasher, rate limiter
   (`infra.go`).
2. Apply the migrations, if `AUTO_MIGRATE` is set (`schema.go`).
3. Check that the schema version matches the latest embedded migration, and
   fail if not (`schema.go`).
4. Build the keyring, which reads the signing keys and adds the first one
   when there is none, then the signer and the token service
   (`infra.go`, `buildSigning`). It comes after the check because it reads
   the database.
5. Build the use cases from the config and the infrastructure
   (`usecases.go`).
6. Build the router (`router.go`), with the limits of the `RATE_LIMIT` level
   from `rate_limit.go`, or none when it is `off`.

A failure at any step after the pool exists closes the pool before
returning, so a startup that fails leaves no open connections behind. Both
`newInfra` and `NewApp` do it with a deferred cleanup guarded by the named
error, rather than a `Close` at each return. On success `App.Close` owns the
pool.

`App.Run` then starts the background tasks and the HTTP server, and blocks
until the server fails or the context is cancelled. On cancellation it lets
in-flight requests finish, for up to 10 seconds. `cmd/api.go` cancels the
context on `SIGINT` or `SIGTERM`.

Every `New` returns an error for a `nil` dependency or an invalid value, so a
wiring mistake stops the service at startup instead of surfacing as a `nil`
dereference on the first request.

The logger is the one exception. `AppParams.Logger` may be `nil`, and `NewApp`
then builds it from the `LOG_FORMAT` in the configuration it was given, which
is the only thing a caller could have built it from anyway: a missing logger is
not a wiring mistake, because nothing has to be guessed. `tests/e2e` takes that
route. From there the logger travels in `dependencies` to the router and to the
background tasks, and no package reads it from `slog.Default()`.

The schema check in steps 2 and 3 is described in
[Migrations](persistence/migrations.md#version-check).

## Background tasks

`periodic.go` runs functions on a ticker, with a timeout for each run, until
the app's context is cancelled. There is one task: every 10 minutes it syncs
the [signing keys](tokens.md#signing-keys). It reads them again, adds the next
one when it is due, and deletes the retired ones.
