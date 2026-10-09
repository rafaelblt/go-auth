# Use cases

`internal/port`, `internal/usecase`

The application logic: what the service does, expressed without HTTP and
without SQL. One package per use case, each depending only on the interfaces
in `internal/port`.

| Document | Covers |
|---|---|
| [Register](register.md) | `/v1/auth/register`: validate, check the username, hash, insert |
| [Login](login.md) | `/v1/auth/login`: verify the credentials, open a session, issue both tokens |
| [Change password](change-password.md) | `/v1/auth/change-password`: verify the credentials, replace the hash, revoke every session |
| [Refresh](refresh.md) | `/v1/auth/refresh`: rotation, reuse detection, client obligations |
| [Verify](verify.md) | `/v1/auth/verify`: check an access token for a service that does not verify it itself |

The tokens these flows hand out are described in [Tokens](../tokens.md).

## Ports

`internal/port` declares the interfaces the inside of the application depends
on, so that nothing in `usecase` or `domain` names a concrete adapter:

| Group | Interfaces |
|---|---|
| Persistence | `UserReader`, `UserWriter`, `UserExistsChecker`, `PasswordReader`, `PasswordWriter`, `SessionReader`, `SessionWriter`, `RefreshTokenReader`, `RefreshTokenWriter`, `UnitOfWork` |
| Passwords | `PasswordHasher`, `PasswordChecker` |
| Access tokens | `AccessTokenIssuer`, `AccessTokenValidator` |
| Keys | `PublicKeyProvider`, `SigningKeyStore` |
| Rate limiting | `RateLimiter` |
| Time | `Clock` |

Three of them are not used by a use case. `PublicKeyProvider` is consumed
directly by the [JWKS handler](../http.md#jwks-is-a-plain-handler), which has
no use case behind it. `RateLimiter` is consumed by the use case adapter,
before any use case runs ([HTTP](../http.md#rate-limiting)).
`SigningKeyStore` is consumed by the keyring in `internal/infra/jwt/ed25519`,
so the PostgreSQL repository implements a port like every other, and no
adapter imports another.

Reading and writing are separate interfaces, even where one type implements
both: `postgres.UserRepo` is a `UserReader`, a `UserWriter` *and* a
`UserExistsChecker`. Each use case then declares only what it uses, and its
`Config` struct says what it can do. `register.Config` has a
`UserExistsChecker` and no `UserReader`: registration cannot read users, and
that is visible without reading its code. `SigningKeyStore` is the exception:
one interface lists, adds and deletes the signing keys, because the keyring,
its only consumer, uses all three.

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
never sends it to the client. This is what lets login, change password,
refresh and verify give every caller the same `invalid_credentials` or
`invalid_token` while telling their failures apart in the logs.

How the HTTP layer turns one into a response:
[Error translation](../http.md#error-translation).
