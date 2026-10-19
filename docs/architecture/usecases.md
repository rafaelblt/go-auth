# Use cases

`internal/port`, `internal/usecase/...`

The application logic: what the service does, without HTTP and without SQL.

## Ports

`internal/port` declares the interfaces the use cases depend on, so that
nothing in `usecase` or the domain names a concrete adapter.

Reading and writing are separate interfaces, even where one type implements
both (`postgres.UserRepo` is a `UserReader`, a `UserWriter` and a
`UserExistsChecker`), so a use case's `Config` says what it can do:
`register.Config` has a `UserExistsChecker` and no `UserReader`.

Three ports have no use case behind them: `PublicKeyProvider`, used by the
JWKS handler; `RateLimiter`, used by the use case adapter before any use case
runs; and `SigningKeyStore`, used by the keyring in
`internal/infra/jwt/ed25519`. They are ports anyway, so that no adapter imports
another.

## The shape of a use case

Each use case is a package under `internal/usecase` with:

- a `Config` holding every dependency, and a `New` that rejects a missing one;
- `Input` and `Output`: plain data, with no HTTP or SQL types;
- one `Execute(ctx, Input) (Output, error)`.

The [session cleanup](#session-cleanup) runs on a timer rather than for a
request, and has nothing to take or give back, so its `Execute` takes only the
context and returns only an error.

`internal/usecase` itself holds what they share: the [DTOs](#dtos) outputs
carry, with their mappers, and `UseCaseError`.

A `UseCaseError` has a `code`, a `kind` (`conflict`, `unauthorized`) and an
optional `reason`. The HTTP layer maps the kind to a status, so `usecase`
never refers to HTTP, and it logs the reason but never sends it
([Error translation](http.md#error-translation)). That is how login, change
password, refresh, verify and me answer every failure with one code while the
log tells the failures apart ([reasons](../logging.md#outcomes)).

### DTOs

A DTO's fields are unexported, with accessors
([Protected struct](conventions.md#protected-struct)), so the only way to build
one that is not zero is its mapper, and every mapper rejects a zero or nil
entity. With exported fields, a use case could build a DTO by hand and skip the
mapper's checks, unnoticed for as long as the result happened to be right. The
HTTP layer checks `IsZero` before it reads a DTO, and panics on a zero one.

That holds because every use case lives in a subpackage, out of reach of the
unexported fields. A use case placed in `internal/usecase` itself could fill
them, and the fields would still look protective, so `tests/docs` checks that
none is there. A mapper guarantees that a DTO came from a valid entity, not from
the right one: mapping the wrong user is for the use case tests to catch.

## Shared credential check

`internal/usecase/credential`

Login and change password take a username and a password, and must reject them
alike, or one endpoint would reveal what the other hides. Both use this
package, which is not a use case:

- `Parse` reads the username and the password with the registration rules
  (`NewUsername`, `NewPlain`). A value that breaks one is rejected like a
  wrong password, never with a `422`, which would say which of the two was
  wrong. These checks fail fast: their time depends only on the input, and the
  rules are public.
- `Checker.Check` finds the user, then its password, and verifies it. When
  either is missing, `rejectWithDummyVerify` first verifies the password
  against a [dummy hash](#the-dummy-hash), so a missing account costs a bcrypt
  comparison like a wrong password does.
- Every rejection is `401 invalid_credentials`, each with its own reason.
  `ErrPasswordChanged` is the one the use cases return themselves, when the
  password is replaced between the check and their write.

### The dummy hash

The same `401` does not take the same time by itself. Returning as soon as the
user or its password is missing would skip bcrypt, which is slow by design, so
an unknown username would answer hundreds of milliseconds sooner than a wrong
password. A miss therefore verifies the submitted password against a dummy hash,
ignores the result, and returns the rejection for that miss: the call is there
for the time it takes, and an error from it can only mean the dummy hash is
broken.

A bcrypt hash carries its cost, and `BCRYPT_COST` is configurable, so
`internal/bootstrap` makes the dummy hash at startup, by hashing a fixed
password with the configured hasher. Login and change password get it in their
`Config`, and pass it to the `Checker` each builds in `New`. Considered and
rejected:

- **A hash written in the code**: its cost is fixed, so any other `BCRYPT_COST`
  reopens the gap, and another hasher would reject it with an error, a `500`
  that reveals more than the timing did.
- **Hashing it on the first miss**: that miss pays two hashes, and the use
  cases gain a dependency used only for this.
- **Padding every response to a fixed duration**: the right duration depends on
  hardware and load, and a slow outlier still leaks.

What is left: a missing user skips the `FindByUserID` query, one database
round trip, small next to bcrypt and network jitter; after a `BCRYPT_COST`
change, the accounts hashed before no longer take as long as a missing one
([Limitations](../limitations.md#login-timing-after-a-bcrypt_cost-change)); and
registration says whether a username is taken anyway
([Limitations](../limitations.md#registration-reveals-whether-a-username-exists)).

## Register

`internal/usecase/register` creates an account. It does not log the user in.

1. Validate the username and the password, collecting every issue: a `422`.
2. `ExistsByUsername`: when taken, `409 username_already_exists`.
3. Hash the password.
4. Build the `User` and its `Password`, with one `now`.
5. In one transaction, insert both. Taken meanwhile: the same `409`.

Step 5 is atomic, or a failure between the inserts would leave a user without
a password: an account nobody can log into, holding its username forever.

Step 2 does not decide whether the username is free, since two concurrent
registrations can both pass it. The insert decides, as a
[guarded write](persistence.md#guarded-writes), and the loser gets the same
`409`. Step 2 stays because it answers the common case with an index lookup,
before a bcrypt hash is spent on a name that is taken.

## Login

`internal/usecase/login` checks a username and a password, opens a session,
and returns both tokens.

1. The [shared credential check](#shared-credential-check).
2. Build the `Session`, issue the access token, and build the first
   `RefreshToken`: no parent, expiring `REFRESH_TOKEN_TTL` from now.
3. In one transaction: `LockHash` on the password step 1 checked, then insert
   the session and the refresh token.

Every rejection is `401 invalid_credentials`, so login never answers `422`,
even to input registration would reject with one.

Step 3 is atomic, so a refresh token never exists without its session.

### The password lock

No session opened with a password may outlive a change of that password. A
change revokes the sessions that exist when it runs, but login checks the
password, which takes hundreds of milliseconds, before it saves the session: a
change that commits in between misses that session, which then refreshes for
good with a password that no longer works. Login writes `sessions` and the
change writes `passwords`, so neither write can carry the other's condition as
a [guarded write](persistence.md#guarded-writes), and under `READ COMMITTED`
neither sees the other's uncommitted row.

So login's transaction first runs `LockHash`, which locks the password row only
while its hash is still the one step 1 checked, and the two meet on that row:

- **The change first.** Its `UPDATE` holds the row, so the lock waits for it to
  commit, then matches nothing. Login rolls back, never returns the access
  token it issued, and answers `401` with the reason `password changed`, as it
  would have a moment later.
- **The login first.** The change's `UPDATE` waits for login to commit, and the
  revocation that follows reaches the session.

A `401` can therefore answer a password that was right when it was sent. Every
login runs one more indexed query, and a change waits, for milliseconds, for the
logins of the same user that are saving their sessions; bcrypt stays outside
the transaction. Any new path that opens a session from a checked password has
to take the same lock. Considered and rejected:

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

`internal/usecase/changepassword` replaces a user's password, given the
username and the current one, and revokes every session of the user. It
returns no token.

1. `credential.Parse` on the username and the current password.
2. Validate the new password: the registration rules, and not the current
   password (`same_as_current`). A failure is a `422` naming `new_password`.
3. `Checker.Check` on the credentials.
4. Hash the new password, and `ChangeHash` the `Password`.
5. In one transaction: `UpdateHash`, only while the stored hash is still the
   one step 3 checked, then `RevokeAllByUserID`.

Step 2 comes before the account is looked up: its answer depends only on the
input and on public rules, so it says nothing about the account, and a request
that cannot succeed costs no bcrypt comparison.

Every session is revoked because a password is most often changed when someone
else may know it, and a session has
[no maximum age](../limitations.md#sessions-have-no-maximum-age): a refresh
token obtained with the old password would keep its holder signed in for good.
The user is signed out everywhere, including where the password was changed.

Two changes with the same current password can both pass step 3. `UpdateHash`
is a [guarded write](persistence.md#guarded-writes), so exactly one applies;
the other rolls back and gets `401` with the reason `password changed`, since
the password it presented is no longer the user's. A login running alongside is
covered by [login's lock](#the-password-lock).

## Refresh

`internal/usecase/refresh` exchanges a refresh token for a new pair, and
carries reuse detection.

1. Parse the token, hash it, `FindByHash`. Malformed or unknown:
   `invalid token`.
2. Load its session. Gone: `invalid token`. The foreign key keeps a token from
   outliving its session, so it is the [session cleanup](#session-cleanup)
   deleting both between the two reads, and the token is unknown from then on.
3. Session revoked: `session revoked`.
4. `token.Use(now)`. Already used, expired or not: reuse detection. Expired:
   `token expired`. Otherwise, rotate.

Every rejection is `401 invalid_token`, with the reason only in the log, so the
response tells someone holding a stolen token nothing about its state.

### Rotation

Every refresh spends the token presented and returns a new one, whose parent is
the spent one and which gets a full `REFRESH_TOKEN_TTL`. A token that stayed
valid until it expired would be worth stealing for that whole time, and a copy
could not be told from the original; rotation limits that, and is what makes
reuse detectable at all.

In one transaction, `MarkUsed` spends the token and the new one is inserted, so
a failure between the two cannot burn the client's only token. `MarkUsed` is a
[guarded write](persistence.md#guarded-writes) on `used_at IS NULL`: two
requests with the same token can both pass step 4, and exactly one rotates.

The others roll back and are answered as [reuse](#reuse-detection): whether a
spent token is found on read or at write, the session is revoked. Two requests
presenting one token at once are what reuse detection describes, two parties
holding it. Arriving together makes neither more trustworthy, and if the
attacker's arrives first, revoking is the only thing that locks them out.
Considered and rejected:

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

A refresh token is used once, so a second presentation means two parties hold
it: the client, and someone who copied it. The service cannot tell which one
is calling, so it revokes the whole session, which locks out both: the attacker
for good, the user until they log in again. Serving the request instead would
let an attacker who replayed a token keep a rotating credential indefinitely.

The session revoked can already be dead: a spent token comes here even once it
has expired (step 4), and the [session cleanup](#session-cleanup) can delete a
dead session after step 2. The revocation then finds no session, and the
answer is `invalid token`, as at step 2: the token is unknown from then on.

The legitimate client pays when it sends a token twice itself, after a lost
response or from two tabs
([Limitations](../limitations.md#a-repeated-refresh-logs-the-user-out)). The
`parent_id` chain keeps the full rotation history of a compromised session.

Revoking a session leaves its access tokens valid until they expire
([Limitations](../limitations.md#access-tokens-cannot-be-revoked)). A refresh
already past step 3 when the session is revoked can still finish, as if it had
finished just before, and the refresh token it returns belongs to the revoked
session.

## Verify

`internal/usecase/verify` checks an access token for a service that would
rather not verify it itself.

It reads the clock, then validates the token with
`infra/jwt.AccessTokenService.Validate`, whose
[checks](tokens.md#access-tokens) say which tokens are `invalid token` and
which `token expired`; both answer `401 invalid_token`, and anything else
`500`.
`ExpiresIn` is never negative, because the clock is read before the token is
checked.

It reads no database: it runs the code a local verifier is pointed to as a
reference ([What to verify](../api/token-verification.md#what-to-verify)), so a
token passes here exactly when it would pass a correct verifier elsewhere, and a
service can switch between the two. An access token names a user, not a
session, so the tokens of a revoked session pass until their `exp`. It is never
rate limited ([why](http.md#rate-limiting)).

## Me

`internal/usecase/me` returns the user an access token was issued to, for a
client that holds the token and needs more than the user ID it carries, such
as the username.

It validates the token as [verify](#verify) does, then reads the user its
`sub` names with `UserReader.FindByID`. A token verify would reject is
`invalid token` or `token expired`, as there, and a token whose user is no
longer stored is `user not found`: all three answer `401 invalid_token`, and
anything else `500`. Nothing in the service deletes a user, so that last one
takes a change made outside it, such as a row deleted by hand or a database
restored from an older backup.

Like verify, it reads nothing about the token's session, which the token does
not name, so the tokens of a revoked session still get their user until their
`exp`. Unlike verify, it is rate limited ([why](http.md#rate-limiting)).

## Session cleanup

`internal/usecase/sessioncleanup` deletes the sessions that have been dead for
longer than [`SESSION_RETENTION`](../configuration.md#session_retention), with
their refresh tokens, so that `sessions` and `refresh_tokens` stop growing for
as long as the deployment runs. A session is dead once it is revoked, or once
its latest refresh token has expired: nothing can refresh it again. It runs on
every instance as a [background task](startup.md#background-tasks), and not at
all while the retention is `off`.

Each run reads the clock once, and deletes the sessions that died before that
time minus the retention ([which ones](persistence.md#deleting-dead-sessions)),
a batch at a time, until a batch comes back short or the run has made its
maximum number of batches ([sizes](startup.md#background-tasks)). A run
therefore deletes a bounded number of sessions, and a backlog, such as the one
a deployment that ran without the cleanup has, takes as many runs as it needs
rather than one long one, though a deployment can fall behind
([Limitations](../limitations.md#the-session-cleanup-can-fall-behind)).

A refresh that presents a token of a session at the moment the cleanup deletes
it, a token dead for longer than the retention and refused either way, gets
`invalid token`, as it would a moment later, whether the session is gone when
the refresh reads it ([Refresh](#refresh), step 2) or when reuse detection
revokes it ([Reuse detection](#reuse-detection)).
