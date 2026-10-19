# Domain

`internal/domain/user`, `internal/domain/password`, `internal/domain/session`,
`internal/validation`, `internal/shared`

Four entities in three packages, the validation types they share, and a few
generic helpers. What the entities are, seen from outside, is in
[Concepts](../system.md#concepts). Every type follows the
[code conventions](conventions.md): fields are unexported, and state changes
only through methods that enforce the type's own rules.

## Entities

Every identity (`user.ID`, `password.ID`, `session.SessionID`,
`session.RefreshTokenID`) wraps `shared.EntityID`, a UUID, version 4 when
generated, that rejects the nil UUID. Functions often take several IDs at once,
and with a bare `uuid.UUID` swapping two of them would compile and fail later as
"not found"; wrapped, it does not compile.

`user` and `password` are separate packages because a user *may* have a
password, so another kind of credential can be added later as its own entity
without changing `User`. `session` holds both `Session` and `RefreshToken`
because they change together: revoking a session invalidates its tokens, and
reuse detection reads a token and revokes its session in one operation. In
separate packages, each would import the other.

**When you change an entity:** its model and mappers in
`internal/infra/postgres/models.go`, a migration if a column changes
([Recipes](recipes.md#add-a-migration)), and its factory in `testutil`.

## `user`

`internal/domain/user`

| File | Holds |
|---|---|
| `user.go` | `User`, `NewUser`, `RestoreUser` |
| `username.go` | `Username`, `NewUsername`, its validators |
| `status.go` | `Status`, `ParseStatus` |
| `id.go` | `ID`, `NewID`, `ParseID` |
| `errors.go` | `ErrUsernameAlreadyExists`, which the writer returns |

**Rules the code keeps**

- `NewUsername` lower-cases, then validates against the
  [username rules](../system.md#usernames-and-passwords), built once from
  constants into `usernameValidators`. A stored username is rebuilt with it.
- `Status` is a closed set with one value, `active`; `ParseStatus` rejects
  anything else, so an unexpected value in the database is an error rather than
  a state nobody handles.

## `password`

`internal/domain/password`

| File | Holds |
|---|---|
| `password.go` | `Password`, `NewPassword`, `RestorePassword`, `ChangeHash` |
| `plain.go` | `Plain`, `NewPlain`, `NewPlainReplacing`, `PlainMinCodePoints` and the other limits |
| `hashed.go` | `Hashed`, `NewHashed` |
| `id.go` | `ID`, `NewID`, `ParseID` |
| `errors.go` | `ErrHashChanged`, which the guarded writes return |

**Rules the code keeps**

- `Password` holds a hash and its user's ID. One password per user is the
  `UNIQUE` on `passwords.user_id`, not the domain's.
- `ChangeHash` replaces the hash and keeps the ID and `createdAt`: a new
  password is the same row with another hash, and the write that persists it is
  [guarded](persistence.md#guarded-writes).
- `NewPlain` applies the [password rules](../system.md#usernames-and-passwords)
  and changes nothing in the value. The validators are built once from
  constants; a fork that needs another minimum changes `PlainMinCodePoints`.
- `NewPlainReplacing(value, current)` applies the same rules, then refuses
  `current` itself, byte for byte, with `same_as_current`.
- `Hashed` comes from a trusted component, so its only rule is that it is not
  empty.

## `session`

`internal/domain/session`

| File | Holds |
|---|---|
| `session.go` | `Session`, `SessionID`, `NewSession`, `RestoreSession`, `Revoke` |
| `refresh_token.go` | `RefreshToken`, `RefreshTokenID`, `NewRefreshToken`, `RestoreRefreshToken`, `Use` |
| `refresh_token_secret.go` | `RefreshTokenSecret`, `ParseRefreshTokenSecret`, `Hash` |
| `refresh_token_hash.go` | `RefreshTokenHash` |
| `access_token.go` | `AccessToken` |
| `errors.go` | the sentinels the use cases branch on: `ErrTokenInvalid`, `ErrTokenExpired`, `ErrTokenAlreadyUsed`, `ErrSessionNotFound` |

**Rules the code keeps**

- **`Revoke` is idempotent** in memory: on a revoked session it keeps the first
  time. `SessionWriter.RevokeAllByUserID` revokes in SQL instead, so it reaches
  sessions nobody loaded, and sets `revoked_at` only where it is still `NULL`.
- **`NewRefreshToken` returns the token and its secret.** The token keeps only
  the secret's hash. Its params carry no hash and the secret's constructor is
  unexported, so a stored hash always comes from a real secret, and the secret,
  being a return value rather than a field, never reaches a writer, a log line
  or a restored token.
- **`Use` checks "already used" before "expired"**
  ([why](../system.md#reuse-detection)). It decides in memory, on a token read
  before the write, so the write checks again
  ([Rotation](usecases.md#rotation)).
- **`RefreshTokenSecret`** is 32 bytes from `crypto/rand`, which `Value()`
  encodes in base64url. `ParseRefreshTokenSecret` reads a client's string back,
  rejecting a bad encoding or length. It is `Parse`, not `Restore`, because the
  secret is never stored. `Value()` and `Hash()` panic on a zero secret: one can
  only come from an ignored error, and its hash would otherwise be looked up as a
  token that simply does not exist, hiding the bug.
- **`RefreshTokenHash`** is exactly 32 bytes. **`AccessToken`** is a signed JWT,
  not empty, typed so it cannot be mixed up with another string on its way out.

## `validation`

`internal/validation`

The types for input rules, with two callers: the domain, whose failures the API
answers with a `422`, and `internal/config`, whose failures become the
[startup error](startup.md#configuration-loading). It is not a domain package,
and a change to it changes both reports.

| File | Holds |
|---|---|
| `issue.go` | `Issue`, `Issues` |
| `issuers.go` | one constructor per issue code |
| `validators.go` | `Validator`, `Validate`, and the validators |
| `accumulator.go` | `Accumulator` |
| `field_error.go`, `validation_error.go` | `FieldError`, `ValidationError` |
| `consts.go` | `LengthUnit` |

**Rules the code keeps**

- An **`Issue`** is one broken rule: a `code` and a `details` map, such as
  `too_short` with `{"min": 3, "unit": "code_point"}`. It does not implement
  `error`: nothing treats one issue as an error, and since `details` is a map, an
  `Issue` is not comparable, so `errors.Is` against one would quietly return
  `false`.
- A **`Validator[T]`** is `func(T) *Issue`, where `nil` means the value passed.
  `Validate(value, validators...)` runs every one and collects every failure.
- `MinLength` and `MaxLength` are separate, so each produces one code; a single
  `Length(min, max)` would take two adjacent `int`s and invite swapping them.
  Every length is counted in an explicit `LengthUnit`, which reaches the API
  response ([why](../system.md#usernames-and-passwords)).
- `internal/config` writes its own validators where it needs one
  (`allowedLogFormat`, `allowedRateLimitLevel`, which return `not_allowed`), and
  `same_as_current` has no validator, since it compares two values.
- The **`Accumulator`** collects the issues of several fields. `Add` returns
  nothing, so there is no `if err != nil` after each field, and `Err()` returns a
  `ValidationError` holding every `FieldError`, or `nil`. This is how
  registration reports a bad username and a bad password in one response, and
  how a startup failure lists every misconfigured variable.

**When you change it:** a new issue code is part of the API: it goes in the
[Error model](../api/errors.md#validation-errors) and the OpenAPI spec.

## `shared`

`internal/shared`

Helpers with no domain meaning: the `EntityID` every identity wraps
(`entity_id.go`), a `Set` (`collections.go`, the allowed username characters),
and pointer helpers for optional fields (`utils.go`).
