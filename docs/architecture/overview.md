# Architecture overview

`go-auth` is a layered application in the ports-and-adapters style. The
business rules sit in the middle and know nothing about HTTP, PostgreSQL,
bcrypt or JWT. Those live at the edges, and the middle reaches them only
through interfaces.

It is a standalone service with an HTTP and JSON API rather than a Go library,
so that projects in any language can use it. The cost is a network round trip
on every call, which is why access tokens can be verified without calling it
([Token model](tokens.md#token-model)). Everything is under `internal/`: the
public interface is HTTP.

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
| `cmd`, `internal/bootstrap` | Entry point, and the composition root: wires everything, runs the server and the background task | [Startup](startup.md) |
| `internal/config` | Loads and validates the environment variables | [Configuration](../configuration.md) |
| `internal/api` | HTTP: routing, JSON bodies, middleware, rate limiting, error translation | [HTTP](http.md) |
| `internal/usecase/...` | One package per use case, `credential`, the check login and change password share, and the DTOs | [Use cases](usecases.md) |
| `internal/port` | The interfaces the use cases depend on | [Use cases](usecases.md#ports) |
| `internal/domain/user`, `password`, `session` | Entities and value objects | [Domain](domain.md) |
| `internal/validation` | Input rules: issues, validators, accumulator | [Domain](domain.md#validation) |
| `internal/shared` | Generic helpers | [Domain](domain.md#shared) |
| `internal/infra` | `SystemClock` | [Conventions](conventions.md#clock) |
| `internal/infra/postgres`, `internal/infra/migrate`, `migrations` | Pool, repositories, unit of work, schema check, migrations | [Persistence](persistence.md) |
| `internal/infra/bcrypt` | Password hashing | [`BCRYPT_COST`](../configuration.md#bcrypt_cost) |
| `internal/infra/jwt`, `internal/infra/jwt/ed25519` | Access tokens and the signing keys | [Tokens](tokens.md) |
| `internal/infra/ratelimit` | The in-memory rate limiter | [HTTP](http.md#rate-limiting) |
| `internal/testutil/...`, `tests/e2e`, `tests/docs` | Test helpers, end-to-end and documentation tests | [Testing](../development/testing.md) |

## A request, end to end

A login:

```
POST /v1/auth/login
  │
  ├─ logging middleware      request ID, logger in the context, "request received"
  ├─ recovery middleware     a panic from here on becomes a 500
  ├─ jsonRouteErrors         the mux's 404 and 405 become JSON
  ├─ ServeMux                match POST /v1/auth/login
  │
  ├─ adaptUseCase            Content-Type not application/json → 415
  │                          rate limit used up → 429
  │                          body past 64 KiB → 413, not JSON → 400
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
