# Code conventions

Patterns that recur throughout the code. Following them keeps new code
consistent with what is already here. The conventions for tests are in
[Testing](../development/testing.md).

## Protected struct

Unexported fields, exported accessor methods.

```go
type User struct {
    id       ID
    username Username
}

func (u *User) ID() ID             { return u.id }
func (u *User) Username() Username { return u.username }
```

Used for every entity, value object, `Config` and DTO. Code outside the
package cannot assign to a field, so the checks in the constructor cannot be
bypassed afterwards. For DTOs the point is how they are built rather than
immutability: use cases live in subpackages of `internal/usecase`, so a mapper
is the only way they can get a DTO that is not zero
([decision 0034](../development/decisions/0034-protected-dtos.md)).

Accessors are named after the field, without a `Get` prefix, as is usual in
Go.

**Where it is not used:** params structs, and use case `Input` and `Output`.
They are plain data, and an `Output` is built in its own package, where
unexported fields would restrict nothing. To decide for a new type: protect it
when there is a rule for how it is built, and code that could break that rule
lives in another package.

## Params struct

Constructors and functions with more than about two arguments take a struct.

```go
type CreationParams struct {
    Username  Username
    CreatedAt time.Time
}

func NewUser(params CreationParams) (*User, error)
```

Every value is named at the call site, so two arguments of the same type
cannot be swapped, and adding a field does not break existing callers.

Names: `CreationParams` and `RestoreParams` for entities, `Config` for
components with dependencies.

## Config struct with a validating constructor

Every component takes its dependencies in a `Config`, and rejects invalid
ones.

```go
type Config struct {
    UserReader      port.UserReader
    Clock           port.Clock
    RefreshTokenTTL time.Duration
}

func New(cfg Config) (Login, error) {
    if cfg.UserReader == nil {
        return Login{}, errors.New("user reader cannot be nil")
    }
    if cfg.Clock == nil {
        return Login{}, errors.New("clock cannot be nil")
    }
    if cfg.RefreshTokenTTL <= 0 {
        return Login{}, errors.New("refresh ttl zero or negative")
    }
    // ...
}
```

The `Config` struct lists exactly what the component needs: `register.Config`
has four fields, so registration touches four things. And a wiring mistake
fails at startup with an error that names it, not with a `nil` dereference on
the first request that takes that path.

## Value objects

Domain values are types, not bare strings and UUIDs. A `user.Username` cannot
be passed where a `password.Plain` is expected, nor a `user.ID` where a
`session.SessionID` is.

### Two constructors

| Constructor | For | Checks |
|---|---|---|
| `NewX(input)` | values arriving from outside | input rules **and** invariants |
| `RestoreX(params)` | values read back from storage | invariants only |

Input rules, such as a minimum length, can reasonably be tightened later.
Invariants, such as a non-nil ID or a non-empty hash, hold for any valid
value. `New` guards the boundary, so it checks both. `Restore` checks only the
invariants, so data already accepted stays readable: if `Restore` applied the
input rules, raising the minimum username length would make existing users
fail to load, and a policy change would need a data migration.

Values read from a string in a fixed format use `ParseX`: `user.ParseID`,
`user.ParseStatus`, `session.ParseRefreshTokenSecret`.

### Two error conventions

| Signature | When |
|---|---|
| `NewX(input) (X, validation.Issues)` | Input from users: `Username`, `Plain` |
| `NewX(input) (X, error)` | Values generated internally: `Hashed`, `AccessToken` |

Returning `Issues` from an input constructor is a contract: its failure is
always a set of broken rules that can be reported, so `Accumulator` takes it
directly. With `error`, every caller would need an `errors.As` to get the
issues back, and a branch for some other error that never happens. A value
generated internally has nothing to report to a user: its failure is a
programming error, so it returns `error`.

### `IsZero`

Every value object has one. Constructors check `IsZero()` on their inputs, so
a zero value that did not come from a constructor is caught instead of spread.

### `Value()` rather than `String()`

Types holding sensitive data expose `Value()`, never `String()`:
`password.Plain`, `password.Hashed`, `session.AccessToken`,
`session.RefreshTokenHash` and `session.RefreshTokenSecret`.

`fmt` and `slog` call `String()` on their own, so a struct holding such a type
would print it through any `%v`, without anyone deciding to log it. `Value()`
has to be called on purpose. `Username` and `Status` do have `String()`: they
are not sensitive, and printing them is useful.

## Defensive copying

A value object holding a slice, a map or a pointer copies it on the way in and
on the way out.

```go
func NewRefreshTokenHash(value []byte) (RefreshTokenHash, error) {
    // ...
    return RefreshTokenHash{value: slices.Clone(value)}, nil
}

func (h RefreshTokenHash) Value() []byte {
    return slices.Clone(h.value)
}
```

Copying a struct copies the slice header, not the array behind it. Without
both clones, whoever passed the slice in, or read it back, could change a
stored hash later: a silent corruption of exactly the data that must not
change. The same applies to `Issue.Details()` (`maps.Clone`) and to nullable
pointer fields (`shared.ClonePtr`).

## Nullable fields

`*T` for the field, and a comma-ok accessor.

```go
type Session struct {
    revokedAt *time.Time
}

func (s *Session) RevokedAt() (time.Time, bool) {
    if s.revokedAt == nil {
        return time.Time{}, false
    }
    return *s.revokedAt, true
}

func (s *Session) IsRevoked() bool { return s.revokedAt != nil }
```

The pointer never leaves the entity, so no caller can change the entity
through it. `shared.PtrFromOk` turns the comma-ok result back into a pointer
when mapping to a database model.

When the question is only whether it is set, a predicate (`IsRevoked`,
`IsUsed`, `HasParent`) reads better than a discarded comma-ok.

## Behaviour on entities

State changes are methods that enforce their own rules, not field assignments
by the caller.

```go
func (t *RefreshToken) Use(usedAt time.Time) error {
    if t.usedAt != nil { return ErrTokenAlreadyUsed }
    if usedAt.After(t.expiresAt) { return ErrTokenExpired }
    t.usedAt = &usedAt
    t.updatedAt = usedAt
    return nil
}
```

The rule lives in one place, can be tested without a database, and cannot be
skipped by a use case.

Idempotent operations are idempotent in the code too: `Session.Revoke` returns
early if the session is already revoked, so a second revocation cannot
overwrite the first timestamp.

## Injected clock

Nothing but `infra.SystemClock` calls `time.Now()`. Components take a
`port.Clock`.

```go
now := uc.clock.Now()
```

Tests replace it with a fake clock, which is how expiry and reuse detection
are tested without sleeping. A flow usually reads the clock once and reuses
`now`, so entities created together share a timestamp.

## Error wrapping

Errors are wrapped with context on the way up, with `%w`:

```go
if err != nil {
    return Output{}, fmt.Errorf("password hashing failed: %w", err)
}
```

Messages describe the operation that failed, in lower case, without
punctuation. `%w` keeps the chain, so `errors.Is` and `errors.As` still work
at the top of the stack, which `translateError` relies on.

Sentinel errors are package-level `var`s (`session.ErrTokenExpired`,
`postgres.ErrSchemaDirty`), compared with `errors.Is`, never by message.

## Naming

| Thing | Convention | Example |
|---|---|---|
| Constructor | `New<Type>` | `NewUsername` |
| Rebuild from storage | `Restore<Type>` | `RestoreUser` |
| Read from a string | `Parse<Type>` | `ParseStatus` |
| Accessor | field name, no `Get` | `Username()` |
| Zero check | `IsZero()` | |
| State predicate | `Is<State>` / `Has<Thing>` | `IsRevoked()`, `HasParent()` |
| Sentinel error | `Err<Condition>` | `ErrTokenExpired` |
| Fake | `Fake<Interface>` | `FakeUserReader` |

Package names are singular nouns (`user`, `password`, `session`,
`validation`), and are part of the type's name where it is used: `user.ID`,
not `user.UserID`.

**The exception is `session`**, the one package holding two entities. Its
`Session` types keep the entity in their names — `session.SessionID`,
`session.SessionCreationParams`, `session.SessionRestoreParams` — so that they
read unambiguously next to `session.RefreshTokenID` and
`session.RefreshTokenCreationParams`. A `session.ID` would not say which of the
two it identified, and the two do appear in one signature:
`RefreshTokenRestoreParams` carries both. A package with one entity uses the
short form: `user.ID`, `password.ID`, `user.CreationParams`.

The rule for a new package, then: drop the package name from the type unless the
package holds more than one entity.

## Formatting

Standard `gofmt`. Before committing:

```bash
gofmt -l .        # list files that need formatting
gofmt -w .        # format them
go vet ./...      # catch suspicious constructs
```
