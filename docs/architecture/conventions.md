# Code conventions

The patterns that recur throughout the code. Following them keeps new code
consistent with what is already here. Test conventions are in
[Testing](../development/testing.md).

## Principles

Simplicity and readability come first. Each pattern below removes a way to get
something wrong; none is a goal in itself, and where one buys nothing, the
simpler code wins.

**Comments are the exception.** Code says what it does through its names and
its shape; if it needs a comment to be understood, first try to make it
simpler. A comment is for what the simplest code still cannot say: a reason, a
constraint, a non-obvious consequence. When that takes more than a few lines,
the explanation goes in a document and the comment links to it
(`See docs/...`). Code that is there for its timing or its atomicity rather
than its result looks removable: give it such a comment, and read what one
links to before changing the code it is on.

## Protected struct

Unexported fields, exported accessors named after the field, with no `Get`
prefix.

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
bypassed afterwards. For the DTOs, what it protects is how they are built
([DTOs](usecases.md#dtos)).

Params structs and use case `Input` and `Output` are plain data with exported
fields: an `Output` is built in its own package, where unexported fields would
restrict nothing. For a new type, protect it when there is a rule for how it is
built and code that could break that rule lives in another package.

## Params struct

A function with more than about two arguments takes a struct, so every value is
named at the call site, two arguments of the same type cannot be swapped, and
adding a field does not break callers. Entities take `CreationParams` and
`RestoreParams`; components take a `Config`.

## Config struct with a validating constructor

Every component takes its dependencies in a `Config`, and its `New` rejects
missing or invalid ones:

```go
func New(cfg Config) (Login, error) {
    if cfg.Clock == nil {
        return Login{}, errors.New("clock cannot be nil")
    }
    if cfg.RefreshTokenTTL <= 0 {
        return Login{}, errors.New("refresh ttl zero or negative")
    }
    // ...
}
```

The `Config` lists exactly what the component needs, and a wiring mistake fails
at startup with an error that names it, not with a `nil` dereference on the
first request that takes that path.

## Constructors

| Constructor | Builds | Returns |
|---|---|---|
| `New<Entity>(CreationParams)` | a new entity: it generates the ID and sets the initial state | `(*E, error)` |
| `Restore<Entity>(RestoreParams)` | an entity read back from storage: it takes every field and checks the invariants, such as a non-zero ID | `(*E, error)` |
| `New<Value>(input)` | a value from a user: `Username`, `Plain` | `(V, validation.Issues)` |
| `New<Value>(input)` | a value generated internally: `Hashed`, `AccessToken`, `RefreshTokenHash` | `(V, error)` |
| `Parse<Value>(string)` | a value in a fixed string format: the IDs, `Status`, `RefreshTokenSecret` | `(V, error)` |

An input constructor returns `validation.Issues` because its failure is always
a set of broken rules to report, which `Accumulator.Add` takes directly. A
value generated internally has nothing to report to a user, so its failure is
a programming error and an `error`.

Values read back from storage go through the same `New` and `Parse`
constructors as input does, so a stored value has to pass today's rules
([Compatibility](../../CONTRIBUTING.md#compatibility)).

## Value objects

Domain values are types, not bare strings and UUIDs: a `user.Username` cannot
be passed where a `password.Plain` is expected.

- **`IsZero`.** Every value object has one, and constructors check it on their
  inputs, so a zero value that did not come from a constructor is caught
  instead of spread.
- **`Value()` rather than `String()`** on anything sensitive: `password.Plain`,
  `password.Hashed`, `session.AccessToken`, `session.RefreshTokenSecret` and
  `session.RefreshTokenHash`. `fmt` and `slog` call `String()` on their own,
  so a struct holding such a type would print it through any `%v`; `Value()`
  has to be called on purpose. `Username` and `Status` do have `String()`.
- **Defensive copies.** A value holding a slice, a map or a pointer copies it
  on the way in and on the way out (`slices.Clone`, `maps.Clone`,
  `shared.ClonePtr`), so whoever passed it in, or read it back, cannot change
  it later.

## Nullable fields

`*T` for the field, and a comma-ok accessor, so the pointer never leaves the
entity:

```go
func (s *Session) RevokedAt() (time.Time, bool)
func (s *Session) IsRevoked() bool
```

When the question is only whether it is set, a predicate (`IsRevoked`,
`IsUsed`, `HasParent`) reads better than a discarded comma-ok.
`shared.PtrFromOk` turns the comma-ok result back into a pointer for a
database model.

## Behaviour on entities

State changes are methods that enforce their own rules, not field assignments
by the caller: `RefreshToken.Use`, `Session.Revoke`, `Password.ChangeHash`. The
rule lives in one place, is tested without a database, and cannot be skipped by
a use case. An idempotent operation is idempotent in the code too:
`Session.Revoke` on a revoked session keeps the first time.

## Clock

Code reads the time from a `port.Clock`. Components take one in their `Config`,
and `infra.SystemClock`, which returns `time.Now().UTC()`, is the
implementation outside tests. Tests set a fake clock, which is how expiry and
reuse detection are tested without sleeping. A use case reads it once and
reuses `now` for the entities it creates, so they share a timestamp.

The background-task runner in `internal/bootstrap/periodic.go` is the
exception: it ticks with `time.NewTicker` and times each run with `time.Now`,
for its log line only, because the clock cannot time it as well. `UTC()` strips
the [monotonic clock reading](https://pkg.go.dev/time#hdr-Monotonic_Clocks),
so the difference of two `SystemClock` readings is taken on the wall clock,
and is off by however far the system clock is stepped in between. Measured from
`time.Now` with `time.Since`, `elapsed` stays exact, which is what lets it tell
a run that timed out ([Background tasks](../logging.md#background-tasks)).

The `logging` middleware makes the other choice: it times each request with its
`port.Clock`, which lets a test fix the `duration` it logs with the fake clock,
and a clock step during a request skews that one line.

## Errors

Errors are wrapped on the way up with `%w` and a lowercase message naming the
operation that failed (`password hashing failed: %w`), so `errors.Is` and
`errors.As` still work at the top, where `translateError` relies on them.
Sentinel errors are package-level `var`s, compared with `errors.Is`, never by
message.

## Logging

Only `internal/api` and `internal/bootstrap` log. The domain, the use cases and
the adapters return errors, and the HTTP boundary logs each failure once, where
the request ID and the outcome are both known: one failed request is one line,
not one per layer it fell through.

- In `internal/api`, take the logger from the request context with
  `loggerFrom(ctx)`, so the line carries the request ID. In `bootstrap`, use
  the `*slog.Logger` the function is given. Pass the logger by hand rather than
  call `slog` directly: `slog.Default()` is only `loggerFrom`'s fallback for a
  context without a logger, which in the service never happens, since
  `logging` wraps every route, and in the HTTP tests means a line nobody
  asserts on.
- The message is a constant lowercase phrase; everything variable is an
  attribute, in `snake_case` (`error`, `user_id`, `duration`). The startup
  lines ending in `...` predate this.
- The level says who has to act: [Levels](../logging.md#levels).
- Never log a password, a token, a hash, a key, or any value from a request
  body. This holds because of the call sites, not a filter:
  [What is never logged](../logging.md#what-is-never-logged).

To assert on a line in a test, use `contextWithLoggedLines` and `loggedLines`
from `internal/api/middleware_test.go`.

## Naming

| Thing | Convention | Example |
|---|---|---|
| Accessor | field name, no `Get` | `Username()` |
| State predicate | `Is<State>` / `Has<Thing>` | `IsRevoked()`, `HasParent()` |
| Sentinel error | `Err<Condition>` | `ErrTokenExpired` |
| Fake | `Fake<Interface>` | `FakeUserReader` |

Package names are singular nouns and part of the type's name where it is used:
`user.ID`, not `user.UserID`. The exception is `session`, the one package with
two entities: its types keep the entity in their names (`session.SessionID`,
`session.RefreshTokenID`), because a `session.ID` would not say which of the
two it identified. A new package drops the package name from its types unless
it holds more than one entity.
