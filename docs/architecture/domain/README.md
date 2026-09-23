# Domain model

Four entities in three packages, plus the validation types they share. Every
type here follows the [code conventions](../conventions.md): fields are
unexported, values are read through methods, and state changes only through
methods that enforce the type's own rules.

| Document | Covers |
|---|---|
| [User](user.md) | `User`, `Username`, `Status` |
| [Password](password.md) | `Password`, `Plain`, `Hashed` |
| [Session](session.md) | `Session`, `RefreshToken`, `RefreshTokenSecret`, `RefreshTokenHash`, `AccessToken` |
| [Validation](validation.md) | `Issue`, validators, `Accumulator` |

## Entities at a glance

```
User ──1:0..1── Password
 │
 └──1:N── Session ──1:N── RefreshToken
                              │
                              └── parent_id ──► RefreshToken  (rotation chain)
```

| Entity         | Package                    | Identity                                         |
| -------------- | -------------------------- | ------------------------------------------------ |
| `User`         | `internal/domain/user`     | `user.ID`                                        |
| `Password`     | `internal/domain/password` | `password.ID`, refers to a `user.ID`             |
| `Session`      | `internal/domain/session`  | `session.SessionID`, refers to a `user.ID`       |
| `RefreshToken` | `internal/domain/session`  | `session.RefreshTokenID`, refers to a `SessionID` |

## Typed identities

Every identity wraps `shared.EntityID`, a UUIDv4 that rejects the nil UUID.
Functions often take several IDs at once, and with a bare `uuid.UUID` or
`string`, swapping two of them compiles and fails later as "not found".
Wrapped, passing a `user.ID` where a `session.SessionID` is expected does not
compile.

## Why three packages

`user` and `password` are separate on purpose: a user *may* have a password,
and other kinds of credential can be added later as their own entities,
without changing `User`. `Password` refers to its user by `user_id`.

`session` holds both `Session` and `RefreshToken`, because they change
together: revoking a session invalidates its tokens, and reuse detection reads
a token and revokes its session in one operation. In separate packages, each
would import the other.

## Shared helpers

`internal/shared`

| Helper             | Purpose                                              |
| ------------------ | ---------------------------------------------------- |
| `EntityID`         | UUIDv4 identity, rejects the nil UUID                |
| `Set[T]`           | A set backed by a map, used for allowed characters   |
| `Ptr(v)`           | The address of a value, for optional fields          |
| `ClonePtr(p)`      | A copy of a pointer's target, or `nil`               |
| `PtrFromOk(v, ok)` | Turns a comma-ok result into a nullable pointer      |
