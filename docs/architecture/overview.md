# Architecture overview

`go-auth` is a layered application in the ports-and-adapters style. The
business rules sit in the middle and know nothing about HTTP, PostgreSQL,
bcrypt or JWT. Those live at the edges, and the middle reaches them only
through interfaces. Everything is under `internal/`: the public interface is
HTTP ([What it is](../system.md#what-it-is)).

## Reading the guides

The documents in `docs/architecture/` say how the code does what the topic
documents describe, and never restate it: a guide links to the behaviour it
implements, and names errors by their code, not by their HTTP status. To work
on a package:

1. read the behaviour in its topic document, which
   [How go-auth works](../system.md) lists, if it is not familiar yet;
2. read the [code conventions](conventions.md) once;
3. read the package's section, which its doc comment names and the
   [table below](#packages) lists. Every section has the package's files, the
   rules its code keeps, what changes with it, and the decisions behind it;
4. for a common change, follow its [recipe](recipes.md).

## Layers

```
        cmd/api.go                      process entry point
             │
     internal/bootstrap                 composition root: builds everything
             │
   ┌─────────┴─────────┐
   │                   │
internal/api      internal/infra        adapters (inbound / outbound)
   │                   │
   └─────────┬─────────┘
             │
    internal/usecase                    application logic
             │
     internal/port                      interfaces the use cases depend on
             │
     internal/domain                    entities and value objects
  internal/validation, shared
```

**The dependency rule:** dependencies point inward. `usecase` imports `port`
and the domain packages, never `api`, `infra`, `bootstrap` or `config`. The
adapters implement the `port` interfaces and are injected at startup, so every
use case is tested against in-memory fakes, with no database and no HTTP.

## Packages

Each package's doc comment names its document too.

| Package | Role | Document |
|---|---|---|
| `cmd`, `internal/bootstrap` | Entry point, and the composition root: wires everything, runs the server and the background tasks | [Startup](startup.md) |
| `internal/config` | Loads and validates the environment variables | [Startup](startup.md#configuration-loading) |
| `internal/api` | HTTP: routing, JSON bodies, middleware, rate limiting, error translation | [HTTP](http.md) |
| `internal/usecase/...` | One package per use case, `credential`, the check login and change password share, and the DTOs | [Use cases](usecases.md) |
| `internal/port` | The interfaces the use cases depend on | [Use cases](usecases.md#ports) |
| `internal/domain/user`, `password`, `session` | Entities and value objects | [Domain](domain.md) |
| `internal/validation` | Input rules: issues, validators, accumulator | [Domain](domain.md#validation) |
| `internal/shared` | Generic helpers | [Domain](domain.md#shared) |
| `internal/infra` | `SystemClock` | [Conventions](conventions.md#clock) |
| `internal/infra/postgres`, `internal/infra/migrate`, `migrations` | Pool, repositories, unit of work, schema check, migrations | [Persistence](persistence.md) |
| `internal/infra/bcrypt` | Password hashing, and the 4 to 31 range of its cost | [`BCRYPT_COST`](../configuration.md#bcrypt_cost) |
| `internal/infra/jwt`, `internal/infra/jwt/ed25519` | Access tokens and the signing keys | [Tokens](tokens.md) |
| `internal/infra/ratelimit` | The in-memory rate limiter | [HTTP](http.md#rate-limiting) |
| `internal/testutil/...`, `tests/e2e`, `tests/docs` | Test helpers, end-to-end and documentation tests | [Testing](../development/testing.md) |

## A request, end to end

A login:

```
POST /v1/auth/login
  │
  ├─ logging middleware      request ID, logger in the context, "request received"
  ├─ recovery middleware     a panic from here on becomes internal_server_error
  ├─ jsonRouteErrors         the mux's plain-text not found and wrong method become JSON
  ├─ ServeMux                match POST /v1/auth/login
  │
  ├─ adaptUseCase            Content-Type not application/json → unsupported_media_type
  │                          rate limit used up → too_many_requests
  │                          body past the limit → request_body_too_large
  │                          not JSON → invalid_json_body
  ├─ loginDecoder            JSON body → login.Input
  ├─ login.Execute           check the credentials, open a session, issue both tokens
  ├─ loginEncoder            login.Output → response body
  ├─ loginSuccessLog         "success login"
  ├─ writeJSON               marshal, set Content-Type, write status and body
  │
  └─ logging middleware      "request finished", with the status and duration
```

On error, `translateError` takes the encoder's place
([Error translation](http.md#error-translation)).
