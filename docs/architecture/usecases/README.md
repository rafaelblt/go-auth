# Use cases

`internal/port`, `internal/usecase`

The application logic: what the service does, expressed without HTTP and
without SQL. One package per use case, each depending only on the interfaces
in `internal/port`.

| Document | Covers |
|---|---|
| [Register](register.md) | `/v1/auth/register`: validate, check the username, hash, insert |
| [Login](login.md) | `/v1/auth/login`: verify the credentials, open a session, issue both tokens |
| [Refresh](refresh.md) | `/v1/auth/refresh`: rotation, reuse detection, client obligations |

The tokens these flows hand out are described in [Tokens](../tokens.md).

## Ports

`internal/port` declares the interfaces the inside of the application depends
on, so that nothing in `usecase` or `domain` names a concrete adapter:

| Group | Interfaces |
|---|---|
| Persistence | `UserReader`, `UserWriter`, `UserExistsChecker`, `PasswordReader`, `PasswordWriter`, `SessionReader`, `SessionWriter`, `RefreshTokenReader`, `RefreshTokenWriter`, `UnitOfWork` |
| Passwords | `PasswordHasher`, `PasswordChecker` |
| Access tokens | `AccessTokenIssuer`, `AccessTokenValidator` |
| Keys | `PublicKeyProvider` |
| Time | `Clock` |

Two of them are not used by a use case. `PublicKeyProvider` is consumed
directly by the [JWKS handler](../http.md#jwks-is-a-plain-handler), which has
no use case behind it. `AccessTokenValidator` is used by nothing yet: it
describes verifying an access token inside the service, for an endpoint that
is not part of v1. `infra/jwt.AccessTokenService` already implements it
([Verifying access tokens](../../api/token-verification.md#what-to-verify)).

Reading and writing are separate interfaces, even where one type implements
both: `postgres.UserRepo` is a `UserReader`, a `UserWriter` *and* a
`UserExistsChecker`. Each use case then declares only what it uses, and its
`Config` struct says what it can do. `register.Config` has a
`UserExistsChecker` and no `UserReader`: registration cannot read users, and
that is visible without reading its code.

`Clock` is a port so that tests control time. Nothing in the domain or the use
cases calls `time.Now()`.

## The shape of a use case

Each use case has its own package under `internal/usecase`, with the same
shape:

- a `Config` struct holding every dependency, and a `New` that rejects any
  `nil`;
- `Input` and `Output` structs: plain data, with no HTTP or SQL types;
- one `Execute(ctx, Input) (Output, error)` method.

`internal/usecase` itself holds what they share: the DTOs that outputs carry,
the mappers that are the only way to build them
([decision 0034](../../development/decisions/0034-protected-dtos.md)), and
`UseCaseError`.

## UseCaseError

`UseCaseError` carries a `code`, a `kind` and an optional `reason`. The `kind`
classifies the failure (`conflict`, `unauthorized`), and the HTTP layer maps
it to a status code, so `internal/usecase` never refers to HTTP. The kinds
line up closely with HTTP status classes, admittedly, but the mapping lives on
the HTTP side. The `reason` is an internal detail: the HTTP layer logs it and
never sends it to the client. This is what lets login and refresh give every
caller the same `INVALID_CREDENTIALS` or `INVALID_TOKEN` while telling their
failures apart in the logs.

How the HTTP layer turns one into a response:
[Error translation](../http.md#error-translation).
