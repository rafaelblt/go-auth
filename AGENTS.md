# AGENTS.md

`go-auth` is a self-hosted authentication service: a standalone HTTP/JSON API
over its own PostgreSQL database, issuing Ed25519 JWT access tokens and rotating,
revocable refresh tokens. It is at v1: register, login and refresh.

**This repository documents its reasoning in `docs/`.** Almost every *why* is
written down there rather than in code comments, so check `docs/` before
inferring intent from the code. Each package's doc comment names the document
that covers it.

## Read before writing code

[`docs/architecture/conventions.md`](docs/architecture/conventions.md) — the
patterns are specific enough that guessing produces code that does not fit:
`New` vs `Restore` constructors, `Value()` instead of `String()` on secrets,
`validation.Issues` from input constructors but `error` from internally generated
values, params structs past two arguments, protected structs.

## The map

| Layer | Document |
|---|---|
| Overview, dependency rule, package map | [`architecture/overview.md`](docs/architecture/overview.md) |
| Domain entities and value objects | [`architecture/domain/README.md`](docs/architecture/domain/README.md) |
| Ports and use cases | [`architecture/usecases/README.md`](docs/architecture/usecases/README.md) |
| HTTP, middleware, error translation | [`architecture/http.md`](docs/architecture/http.md) |
| Schema, migrations, repositories, guarded writes | [`architecture/persistence/README.md`](docs/architecture/persistence/README.md) |
| Composition root, background tasks | [`architecture/startup.md`](docs/architecture/startup.md) |
| The two token types and the signing keys | [`architecture/tokens.md`](docs/architecture/tokens.md) |
| Every log line the service writes | [`architecture/logging.md`](docs/architecture/logging.md) |
| Endpoints, error model, token verification, OpenAPI spec | [`api/`](docs/api/reference.md) |
| Configuration | [`configuration.md`](docs/configuration.md) |
| What it deliberately does not do | [`limitations.md`](docs/limitations.md) |

## Rules a single file does not reveal

- **v1 is a published contract; keep it backward compatible.** Client
  applications and the services that verify tokens depend on the `/v1`
  endpoints (paths, status codes, error codes, request and response fields, as
  in [`api/`](docs/api/reference.md) and `openapi.yaml`), on the access token's
  header and claims, on the JWKS, and on the configuration variables. Adding is
  compatible: a new endpoint, a new response field, a setting with a default.
  Removing, renaming, or changing the type or meaning of any of these is not.
  Data already stored must keep loading, which is why `Restore` skips input
  rules, and a released migration is never edited: a change is a new one. When
  a change cannot avoid breaking the contract, stop and ask rather than work
  around it.
- **Simplicity and readability come first.** The patterns serve that, not the
  other way round, and a comment is for what the simplest code cannot say. See
  [Principles](docs/architecture/conventions.md#principles).
- **Dependencies point inward.** `usecase` imports `port` and the domain
  packages, never `api`, `infra`, `bootstrap` or `config`.
- **Only `infra.SystemClock` calls `time.Now`.** Everything else takes a
  `port.Clock`.
- **Only `internal/api`, `internal/bootstrap` and `cmd` log.** Do not pass a
  logger into a use case, the domain or an adapter: they return errors, and the
  HTTP boundary logs each failure once. See
  [Logging](docs/architecture/logging.md#where-logging-happens).
- **Never log a password, a token or a hash.** This holds because of the call
  sites, not a filter.
- **Some code exists for its timing or its atomicity, not its result.** Two
  examples that look removable and are not:
  `Login.rejectWithDummyVerify` discards a bcrypt result on purpose, and
  `RefreshTokenRepo.MarkUsed` is guarded in SQL rather than in Go. Both carry a
  `See docs/development/decisions/...` comment; read it before touching them.

## Commands

```bash
go test ./...          # everything; needs Docker
go test $(go list ./... | grep -vE 'infra/postgres|infra/migrate|testutil|tests/e2e')   # fast layers only
gofmt -l . && go vet ./...
```

`-short` skips nothing. Test layers and helpers:
[`development/testing.md`](docs/development/testing.md).

## Commits

Format, types and scopes: [`development/commits.md`](docs/development/commits.md).
**Only commit when asked, and never push.**

## Decision records

[`docs/development/decisions/`](docs/development/decisions/README.md) holds the
seven decisions whose reasoning cannot be read from the code. Its README says when
a new one is warranted; most changes apply an existing decision instead. Each
record's **Invariant** line is what a change must not break — the README table
lists all seven in one place.
