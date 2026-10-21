# Use cases

`internal/port`, `internal/usecase/...`

The application logic, without HTTP and without SQL. What each use case does,
seen from outside, is in [How go-auth works](../system.md#processes); this guide
covers how the code does it. Every section has the same parts: what the package
is, its files, the rules its code keeps, what changes with it, and the decisions
behind it.

## Ports

`internal/port` declares the interfaces the use cases depend on, so that
nothing in `usecase` or the domain names a concrete adapter.

| File | Holds |
|---|---|
| `db.go` | `UnitOfWork` and `UowDeps`; the readers, writers, `UserExistsChecker` and `SessionCleaner` |
| `access_token.go` | `AccessTokenIssuer`, `AccessTokenValidator` and their payloads |
| `password.go` | `PasswordHasher`, `PasswordChecker` |
| `keys.go` | `PublicKeyProvider`, `SigningKeyStore` |
| `rate_limit.go` | `RateLimiter` |
| `clock.go` | `Clock` |

**Rules the code keeps**

- Reading and writing are separate interfaces, even where one type implements
  both (`postgres.UserRepo` is a `UserReader`, a `UserWriter` and a
  `UserExistsChecker`), so a use case's `Config` says what it can do:
  `register.Config` has a `UserExistsChecker` and no `UserReader`.
- A sentinel a method returns for stored state, such as
  `session.ErrSessionAlreadyRevoked` from `SessionWriter.Revoke`, is documented
  on the method ([Guarded writes](persistence.md#guarded-writes)).
- Three ports have no use case behind them: `PublicKeyProvider`, used by the
  JWKS handler; `RateLimiter`, used by the use case adapter before any use case
  runs; and `SigningKeyStore`, used by the keyring. They are ports anyway, so
  that no adapter imports another.

**When you change it:** every implementation and its fake in `porttest`
([Fakes](../development/testing.md#fakes)).

## The shape of a use case

Each use case is a package under `internal/usecase` with:

- a `Config` holding every dependency, and a `New` that rejects a missing one
  ([Config struct](conventions.md#config-struct-with-a-validating-constructor)),
  in `config.go`;
- `Input` and `Output`: plain data, with no HTTP or SQL types;
- one `Execute(ctx, Input) (Output, error)`;
- its rejections as package-level `UseCaseError`s.

The [session cleanup](#session-cleanup) runs on a timer rather than for a
request, and has nothing to take or give back, so its `Execute` takes only the
context and returns only an error.

`internal/usecase` itself holds what they share: the [DTOs](#dtos) outputs
carry, with their mappers (`dtos.go`), and `UseCaseError` (`error.go`).

A `UseCaseError` has a `code`, a `kind` (`conflict`, `unauthorized`) and an
optional `reason`, built with `NewError` or `NewErrorWithReason`. The HTTP layer
maps the kind to a status, so `usecase` never refers to HTTP, and it logs the
reason but never sends it ([Error translation](http.md#error-translation)). One
code with several reasons is how a use case answers
[vaguely](../system.md#deliberately-vague-answers).

**When you change it:** a new reason goes in the reasons table of
[Logging](../logging.md#outcomes); a new code in the [Error model](../api/errors.md#codes)
and the OpenAPI spec; a new kind in the catalogs of
[Error translation](http.md#error-translation).

### DTOs

A DTO's fields are unexported, with accessors
([Protected struct](conventions.md#protected-struct)), so the only way to build
one that is not zero is its mapper, and every mapper panics on a zero or nil
entity. With exported fields, a use case could build a DTO by hand and skip the
mapper's checks, unnoticed for as long as the result happened to be right. The
HTTP layer checks `IsZero` before it reads a DTO, and panics on a zero one.

That holds because every use case lives in a subpackage, out of reach of the
unexported fields. A use case placed in `internal/usecase` itself could fill
them, and the fields would still look protective, so `tests/docs` checks that
none is there. A mapper guarantees that a DTO came from a valid entity, not from
the right one: mapping the wrong user is for the use case tests to catch.

## Shared credential check

`internal/usecase/credential` is not a use case: it is the check login and
change password share, so that they reject alike. Implements the credential
checks of [Login](../system.md#login).

| File | Holds |
|---|---|
| `credential.go` | `Parse`, `Checker.Check`, `rejectWithDummyVerify`, the six `invalid_credentials` errors |
| `config.go` | `Config`, `NewChecker` |

**Rules the code keeps**

- `Parse` reads the username and the password with the registration rules
  (`NewUsername`, `NewPlain`) and answers a value that breaks one like a wrong
  password, never with a `422`. These checks fail fast: their time depends only
  on the input, and the rules are public.
- `Checker.Check` finds the user, then its password, and verifies it. When
  either is missing, `rejectWithDummyVerify` verifies the password against the
  [dummy hash](#the-dummy-hash) first, and returns the rejection for that miss.
- `ErrPasswordChanged` is the one rejection the use cases return themselves,
  when the password is replaced between the check and their write.

**When you change it:** both login and change password, and their tests. A new
path that opens a session from a checked password must take the
[password lock](#the-password-lock).

### The dummy hash

A bcrypt hash carries its cost, and `BCRYPT_COST` is configurable, so
`internal/bootstrap` makes the dummy hash at startup (`newDummyPasswordHash`),
by hashing a fixed password with the configured hasher. Login and change
password get it in their `Config`, and pass it to the `Checker` each builds in
`New`. The result of verifying against it is discarded: the call is there for
the time it takes, and an error from it can only mean the dummy hash is broken.
Considered and rejected:

- **A hash written in the code**: its cost is fixed, so any other `BCRYPT_COST`
  reopens the gap, and another hasher would reject it with an error, a `500`
  that reveals more than the timing did.
- **Hashing it on the first miss**: that miss pays two hashes, and the use
  cases gain a dependency used only for this.
- **Padding every response to a fixed duration**: the right duration depends on
  hardware and load, and a slow outlier still leaks.

## Register

`internal/usecase/register`. Implements [Registration](../system.md#registration).

| File | Holds |
|---|---|
| `register.go` | `Execute`, `ErrUsernameAlreadyExists`, and the steps: `checkUsernameExists`, `createUser`, `createPassword`, `save` |
| `input.go`, `output.go` | `Input`, with `FieldUsername` and `FieldPassword`, the names its validation errors carry; `Output` |
| `config.go` | `Config`, `New` |

**Rules the code keeps**

- Both fields go into one `validation.Accumulator`, so both are reported
  together.
- `checkUsernameExists` decides nothing; `save` inserts both in one unit of
  work, and `UserWriter.Add` is the [guarded write](persistence.md#guarded-writes)
  that decides a taken username, with `user.ErrUsernameAlreadyExists`.
- The `User` and its `Password` share one `now`.

**When you change it:** a newly validated field needs its entry in
`errorFieldCatalog`, which maps those names to the JSON ones
([Error translation](http.md#error-translation)).

## Login

`internal/usecase/login`. Implements [Login](../system.md#login).

| File | Holds |
|---|---|
| `login.go` | `Input`, `Output`, `Execute` |
| `config.go` | `Config`, `New`, which builds the `Checker` |

**Rules the code keeps**

- `credential.Parse`, then `Checker.Check`, then the session, the access token
  and the first `RefreshToken`, with no parent.
- One unit of work runs [`LockHash`](#the-password-lock) on the password just
  checked, then inserts the session and the token. `password.ErrHashChanged`
  from the lock becomes `ErrPasswordChanged`, and the access token already
  issued is dropped.

### The password lock

Login writes `sessions` and a password change writes `passwords`, so neither
write can carry the other's condition as a guarded write, and under
`READ COMMITTED` neither sees the other's uncommitted row. So login's
transaction first runs `PasswordWriter.LockHash`, which locks the password row
only while its hash is still the one step 1 checked, and the two meet on that
row:

- **The change first.** Its `UPDATE` holds the row, so the lock waits for it to
  commit, then matches nothing, and login rolls back.
- **The login first.** The change's `UPDATE` waits for login to commit, and the
  revocation that follows reaches the session.

Every login runs one more indexed query, and a change waits, for milliseconds,
for the logins of the same user that are saving their sessions; bcrypt stays
outside the transaction. Considered and rejected:

- **Reading the password inside the transaction**, `FOR UPDATE` or under
  `SERIALIZABLE`: either bcrypt runs inside it, holding a connection and a lock
  for hundreds of milliseconds per login, or the password is read twice, and
  the second read is this lock.
- **A guarded session insert**, made only while the stored hash is unchanged:
  with both transactions in flight, the insert does not see the uncommitted
  change, nor the revocation the uncommitted session. Tried against two
  connections, it left a session active.
- **A security stamp**, a version on the user copied into each session and
  compared on every refresh: no lock, but a migration, a column on `sessions`
  and a check in refresh. It is the better tool if the service ever needs to
  sign a user out everywhere for other reasons.

## Change password

`internal/usecase/changepassword`. Implements
[Changing the password](../system.md#changing-the-password).

| File | Holds |
|---|---|
| `change_password.go` | `Input`, `Output`, `FieldNewPassword`, `Execute`, `save` |
| `config.go` | `Config`, `New`, which builds the `Checker` |

**Rules the code keeps**

- The new password goes through `password.NewPlainReplacing`, into a
  `ValidationError` on `FieldNewPassword`, before `credential.Parse` and
  `Checker.Check` touch the account.
- `save` runs, in one unit of work, `UpdateHash`, a
  [guarded write](persistence.md#guarded-writes) on the hash `Check` verified,
  then `RevokeAllByUserID`. `password.ErrHashChanged` becomes
  `ErrPasswordChanged`.

A login running alongside is covered by [the password lock](#the-password-lock).

## Refresh

`internal/usecase/refresh`. Implements [Refreshing](../system.md#refreshing).

| File | Holds |
|---|---|
| `refresh.go` | `Input`, `Output`, the four `invalid_token` errors, `Execute`, and the steps: `resolveToken`, `getSessionOfToken`, `handleTokenReuse`, `rotate` |
| `config.go` | `Config`, `New` |

**Rules the code keeps**

- `Execute` takes the steps of [Refreshing](../system.md#refreshing) in their
  order: `resolveToken` (parse, hash, `FindByHash`), `getSessionOfToken`, the
  `IsRevoked` check, then `token.Use(now)`, whose `session.ErrTokenAlreadyUsed`
  goes to [reuse detection](#reuse-detection).
- A session missing in `getSessionOfToken` is `ErrTokenInvalid`, not a bug: the
  foreign key keeps a token from outliving its session, so the session cleanup
  deleted both between the two reads.

**When you change it:** the steps in [Refreshing](../system.md#refreshing),
each error's reason in [Logging](../logging.md#outcomes), and
[logout](#logout), whose `resolveToken` and `getSessionOfToken` are copies of
refresh's.

### Rotation

`rotate` runs, in one unit of work, `MarkUsed` on the token presented and the
insert of its successor, so a failure between the two cannot burn the client's
only token. `MarkUsed` is a [guarded write](persistence.md#guarded-writes) on
`used_at IS NULL`: two requests with the same token can both pass step 4, and
exactly one rotates. The others get `session.ErrTokenAlreadyUsed` from it, roll
back, and go to reuse detection, as a replay found on read does. Considered and
rejected:

- **The `UNIQUE` on `parent_id` alone**: the second insert fails, so the loser
  answers `500` and the session stays active. The constraint stays, so the
  chain cannot fork.
- **A plain `401` for the loser**: it would spare a client that refreshes from
  two tabs at once, but when the legitimate client loses and does not retry, an
  attacker who won keeps a rotating credential.
- **Optimistic locking with a `version` column**: generic, but it needs a
  migration and puts a persistence detail on the entities; worth it only if
  more entities need concurrency control.

### Reuse detection

`handleTokenReuse` revokes the session with `Session.Revoke` and
`SessionWriter.Revoke`, then answers `ErrTokenAlreadyUsed`. `Revoke` is a
[guarded write](persistence.md#guarded-writes) on `revoked_at IS NULL`: a
logout, a password change or another reuse detection can revoke the session
after step 3 read it, and `Revoke` then returns
`session.ErrSessionAlreadyRevoked`, answered with `ErrSessionRevoked`, as at
step 3. A spent token reaches here however long ago it expired, so its session
may be dead and deleted by the session cleanup after step 2: `Revoke` then
returns `session.ErrSessionNotFound`, answered with `ErrTokenInvalid`, as at
step 2. Either way, a request that loses a race gets the answer it would have
got reading a moment later.

Considered and rejected: answering a session revoked first with
`ErrTokenAlreadyUsed`, since the token was reused all the same. The request
would log a revocation it did not make, and a spent token would be logged as
reuse or not depending on milliseconds, where step 3 already lets a revoked
session win over a spent token.

## Logout

`internal/usecase/logout`. Implements [Logging out](../system.md#logging-out).

| File | Holds |
|---|---|
| `logout.go` | `Input`, `Output`, `ErrTokenInvalid`, `Execute`, and the steps: `resolveToken`, `getSessionOfToken`, `revoke` |
| `config.go` | `Config`, `New` |

**Rules the code keeps**

- `resolveToken` and `getSessionOfToken` find the token and its session as
  [refresh](#refresh)'s do, and answer a token or a session that is missing
  with `ErrTokenInvalid`.
- A revoked session is left as it is. Any other is revoked with
  `Session.Revoke` and `SessionWriter.Revoke`, in a unit of work, whatever the
  token's state. `revoke` answers a lost race as
  [reuse detection](#reuse-detection) does: `session.ErrSessionAlreadyRevoked`
  as a session read revoked, and `session.ErrSessionNotFound` as
  `ErrTokenInvalid`.
- The token is never written. Refresh checks the session before the token, so
  the revocation stops every token of the session, and a spent token stays one
  for reuse detection to find.
- `Output.ReuseDetected` is set for a spent token of a session the logout
  revoked, and `Output.SessionAlreadyRevoked` for a session it left as it is.
  `internal/api` logs them with refresh's reasons, for reuse and for a revoked
  session ([Outcomes](../logging.md#outcomes)). The token's state is read for
  nothing else: an expired token ends its session like a current one.

A logout and a rotation of the same session never wait on each other: the
`UPDATE` locks the session row `FOR NO KEY UPDATE`, which does not conflict
with the `FOR KEY SHARE` a rotation's insert takes through the foreign key.
Revocations of the same session do wait on each other, and the one that waited
then matches nothing. A refresh that spends the token between the logout's
read and its write is not reported as reuse: the logout reads the token once
and never writes it, and the session ends revoked either way.

**When you change it:** the steps in [Logging out](../system.md#logging-out),
and the `success logout` line in [Logging](../logging.md#outcomes).

### One revoked state

Logout revokes the session as reuse detection and the password change do,
rather than with a state of its own, such as a `logged_out_at`: `revoked_at`
means the session issues no more tokens, whatever the cause, and everything
that reads it, step 3 of the refresh, the
[session cleanup](persistence.md#deleting-dead-sessions) and
`RevokeAllByUserID`, treats every cause alike. A second state would have each
of them check both. The cause is in the log, not in the database.

## Verify

`internal/usecase/verify`. Implements
[Checking an access token](../system.md#checking-an-access-token).

| File | Holds |
|---|---|
| `verify.go` | `Input`, `Output`, `ErrTokenInvalid`, `ErrTokenExpired`, `Execute` |
| `config.go` | `Config`, `New` |

**Rules the code keeps**

- It reads the clock, then calls `port.AccessTokenValidator.Validate`, and maps
  `session.ErrTokenInvalid` and `session.ErrTokenExpired` to its own errors;
  anything else is unexpected, a `500`. Which tokens are which is decided by the
  validator ([Access tokens](tokens.md#access-tokens)).
- `ExpiresIn` is never negative, because the clock is read before the token is
  checked.
- It has no other dependency: no database.

## Me

`internal/usecase/me`. Implements [Looking up the user](../system.md#looking-up-the-user).

| File | Holds |
|---|---|
| `me.go` | `Input`, `Output`, `ErrTokenInvalid`, `ErrTokenExpired`, `ErrUserNotFound`, `Execute` |
| `config.go` | `Config`, `New` |

**Rules the code keeps**

- It validates the token as [verify](#verify) does, with the same mapping, then
  reads the user with `UserReader.FindByID`. Not found: `ErrUserNotFound`.
- It reads nothing about the session: the token does not name one.

## Session cleanup

`internal/usecase/sessioncleanup`. Implements
[Session cleanup](../system.md#session-cleanup).

| File | Holds |
|---|---|
| `session_cleanup.go` | `Execute` |
| `config.go` | `Config`, `New`: the retention, the batch size and the maximum number of batches |

**Rules the code keeps**

- Each run reads the clock once, and calls `SessionCleaner.DeleteDeadBefore`
  with `now` minus the retention, a batch at a time, until a batch comes back
  short or it has made `MaxBatches` calls. Which sessions count as dead is the
  statement's: [Deleting dead sessions](persistence.md#deleting-dead-sessions).
- `internal/bootstrap` sets the numbers and registers it as a
  [background task](startup.md#background-tasks), and builds none while the
  retention is `off`.
