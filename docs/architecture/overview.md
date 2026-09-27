# Architecture overview

`go-auth` is a layered application in the ports-and-adapters style. The
business rules sit in the middle and know nothing about HTTP, PostgreSQL,
bcrypt or JWT. Those live at the edges, and the middle reaches them only
through interfaces.

## A service, not a library

`go-auth` is meant to be reused by projects written in any language, so it is
a standalone service with an HTTP and JSON API rather than a Go library, which
would serve only Go projects.

The cost is a network round trip on every call to it. That is why access
tokens can be verified without calling it: the round trip happens on login and
refresh, which are rare, and not on the requests an application serves (see
[Token model](tokens.md#token-model)).

Everything is under `internal/`, so no other module can import it. The public
interface is HTTP.

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
    internal/usecase                    application logic: register, login, refresh
             │
     internal/port                      interfaces the use cases depend on
             │
     internal/domain                    entities and value objects
  internal/validation, shared
```

**The dependency rule:** arrows point inward only. `usecase` imports `port`
and the domain packages, and never `api`, `infra`, `bootstrap` or `config`.
The infrastructure implements the `port` interfaces and is injected at
startup.

The immediate payoff is test speed: every use case is tested against
in-memory fakes, with no database and no HTTP, in milliseconds. The longer
term payoff is that replacing bcrypt with argon2, say, means a new package
that implements an existing interface, not a rewrite.

Each layer has its own document:

| Layer | Document |
|---|---|
| `internal/domain`, `validation`, `shared` | [Domain](domain/README.md) |
| `internal/port`, `internal/usecase` | [Use cases](usecases/README.md) |
| `internal/api` | [HTTP](http.md) |
| `internal/infra/postgres`, `migrations` | [Persistence](persistence/README.md) |
| `internal/bootstrap`, `cmd` | [Startup](startup.md) |

## Package map

| Package | Role |
|---|---|
| `cmd` | `main`: load the config, build the app, handle signals |
| `internal/bootstrap` | Composition root: wires every dependency, runs the server and the background tasks |
| `internal/config` | Loads and validates the environment variables |
| `internal/api` | HTTP: routing, JSON bodies, middleware, error translation |
| `internal/usecase` | Application logic, one package per use case |
| `internal/port` | Interfaces the use cases depend on |
| `internal/domain/user` | `User` entity, `ID`, `Username`, `Status` |
| `internal/domain/password` | `Password` entity, `ID`, `Plain`, `Hashed` |
| `internal/domain/session` | `Session` and `RefreshToken` entities, `RefreshTokenSecret`, `RefreshTokenHash`, `AccessToken` |
| `internal/validation` | Input validation: issues, validators, accumulator |
| `internal/shared` | Helpers used across packages: `EntityID`, `Set`, `Ptr`, `ClonePtr` |
| `internal/infra` | `SystemClock` |
| `internal/infra/postgres` | Repositories, connection pool, unit of work, schema version |
| `internal/infra/migrate` | Runs the embedded migrations with golang-migrate |
| `internal/infra/bcrypt` | Password hashing and verification |
| `internal/infra/jwt` | Issues and validates access tokens |
| `internal/infra/jwt/ed25519` | Ed25519 keys, keyring, key store, JWT signer |
| `migrations` | Embedded `.sql` files, and `Latest()` |
| `internal/testutil` | Test helpers, shared across packages |
| `tests/e2e` | End-to-end tests against a real server and database |

## Request lifecycle

A login request, end to end:

```
POST /v1/auth/login
  │
  ├─ logging middleware      give the request an ID, tag the logger, log "request received"
  ├─ recovery middleware     catch panics from here on
  ├─ ServeMux                match POST /v1/auth/login
  │                          no route → 404, wrong method → 405 (JSON, jsonRouteErrors)
  │
  ├─ adaptUseCase            Content-Type not application/json → 415
  │                          decoder reads at most 64 KiB; past it → 413
  ├─ loginDecoder            JSON body → login.Input
  ├─ login.Execute           validate → find user → find password → verify
  │                          → create session → issue access token
  │                          → create refresh token → save both in one transaction
  ├─ loginEncoder            login.Output → response body
  ├─ loginSuccessLog         log "success login"
  ├─ writeJSON               marshal, set Content-Type, write status and body
  │
  └─ logging middleware      log "request finished" with the status and duration
```

On error, `translateError` takes the encoder's place; see
[HTTP](http.md#error-translation).

## Where to look next

- [Domain](domain/README.md): what the entities guarantee.
- [Use cases](usecases/README.md): the three flows in detail.
- [Tokens](tokens.md): the two token types and the security properties.
- [Persistence](persistence/README.md): schema, repositories, transactions.
- [Code conventions](conventions.md): the patterns repeated throughout.
