# Recipes

The changes that come up again and again, with every file they touch, code and
documents, in the order that keeps the build passing. Each step links to the
guide that explains it; this page only says where to go. When a recipe and the
code disagree, the code wins: fix the recipe in the same commit.

## Add an endpoint

For a `POST` backed by a use case. The `me` endpoint (`19bb669`) followed it.

1. **The use case**: a package `internal/usecase/<name>` with `config.go`
   (`Config`, `New`), `<name>.go` (`Input`, `Output`, its errors, `Execute`),
   and a `helper_test.go` with the fakes
   ([The shape of a use case](usecases.md#the-shape-of-a-use-case)). A port it
   needs that has no fake, or a fake missing a capability, goes in
   `internal/testutil/porttest`.
2. **The HTTP pieces**, in `internal/api/endpoints.go`: the request and response
   bodies, a decoder, an encoder and a success log
   ([Use case adapter](http.md#use-case-adapter)).
3. **The route**, in `internal/api/router.go`: a field in `Dependencies`, its
   `nil` check in `NewRouter`, the `adaptUseCase` call and the `mux.HandleFunc`.
   For a rate limit, a field in `RateLimiting`, its check, and an
   `endpointRateLimit` ([Rate limiting](http.md#rate-limiting)).
4. **The wiring**, in `internal/bootstrap`: a `build<Name>` and a field in
   `usecases` (`usecases.go`), the field passed in `newRouter` (`router.go`),
   and for a limit, `rateLimits` and every level of `rateLimitsByLevel`
   (`rate_limit.go`, tested in `rate_limit_test.go`).
5. **Tests**: the use case's own; `internal/api/endpoints_test.go` and
   `router_test.go`; `tests/e2e/<name>_test.go`, and the endpoint added to the
   tables of `content_type_test.go` and `request_body_size_test.go`.
6. **The contract**: a section in [API reference](../api/reference.md) and the
   path in [openapi.yaml](../api/openapi.yaml), which `tests/docs` checks
   against the router; a new error code in the [Error model](../api/errors.md#codes),
   and the endpoint in its client handling block.
7. **The other documents**: the process in
   [How go-auth works](../system.md#processes), and the endpoint among the
   [deliberately vague answers](../system.md#deliberately-vague-answers) if it
   answers several failures with one code; a section in the
   [Use cases](usecases.md) guide; its success line and reasons in
   [Logging](../logging.md#outcomes); its row in
   [`RATE_LIMIT`](../configuration.md#rate_limit) if it is limited; and the list
   of endpoints in the [README](../../README.md).

## Add a reason or an error code

1. The `UseCaseError` in the use case's package, with `NewErrorWithReason` for a
   reason ([The shape of a use case](usecases.md#the-shape-of-a-use-case)).
2. A new `ErrorKind`: its entries in the catalogs of `internal/api/errors.go`
   ([Error translation](http.md#error-translation)).
3. A reason: its row in the reasons table of [Logging](../logging.md#outcomes),
   and the failure in the process in [How go-auth works](../system.md#processes).
4. A code: the [Error model](../api/errors.md#codes), the endpoint's errors in
   the [API reference](../api/reference.md) and the OpenAPI spec. A new code is
   compatible; a changed one is not ([Compatibility](../../CONTRIBUTING.md#compatibility)).

## Add a configuration variable

1. `internal/config`: an `env[T]` in `env_vars.go`, a parser in `parsers.go` if
   none fits, the setting in `config.go` (a `field…` name, the `Config` field
   and accessor, `ConfigParams`, the default, `NewConfig` and `LoadConfig`), the
   name's entry in `envKeyByField` (`env_load.go`), and its validators
   ([Configuration loading](startup.md#configuration-loading)).
2. `internal/bootstrap`: pass the value to the component that needs it.
3. [`.env.example`](../../.env.example), which `tests/docs` checks against
   `env_vars.go`.
4. [Configuration](../configuration.md): a row in the table and a section.
   Optional, with a default that keeps today's behaviour, or it breaks
   [Compatibility](../../CONTRIBUTING.md#compatibility).

## Add a migration

1. The `up` and `down` pair in `migrations/`, with the next number. Check that
   the `down` undoes the `up` on the right table: a wrong `down` only shows up
   during a rollback. `Latest()` and the test helpers pick it up
   ([Migrations](persistence.md#migrations)).
2. A statement that cannot run in a transaction, such as
   `CREATE INDEX CONCURRENTLY`, goes alone in its file.
3. The models and mappers in `internal/infra/postgres/models.go`, and the
   repositories that read or write the column.
4. [Data](../system.md#data) if what is stored changes, and the schema notes in
   [Persistence](persistence.md#schema).

A released migration is never edited: a change is a new migration.

## Add a log line

1. In `internal/api` or `internal/bootstrap` only, with a constant lowercase
   message and `snake_case` attributes, at the level that says who acts
   ([Logging](conventions.md#logging)).
2. Its row in [Logging](../logging.md#every-line), with its fields, and nothing
   from [What is never logged](../logging.md#what-is-never-logged).

## Add a background task

1. Its run function, which takes a context and returns an error, in the
   component that owns the work: a use case for application logic.
2. A `periodicTask` in `App.backgroundTasks` (`internal/bootstrap/app.go`), with
   a name, an interval and a timeout; the tasks registered are checked in
   `app_test.go`.
3. Its row in [Background tasks](startup.md#background-tasks), and its `task`
   name in [Logging](../logging.md#background-tasks).
