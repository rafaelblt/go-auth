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

`internal/usecase` itself holds what they share: the DTOs outputs carry, the
mappers that are the only way to build one
([decision 0034](../development/decisions/0034-protected-dtos.md)), and
`UseCaseError`.

A `UseCaseError` has a `code`, a `kind` (`conflict`, `unauthorized`) and an
optional `reason`. The HTTP layer maps the kind to a status, so `usecase`
never refers to HTTP, and it logs the reason but never sends it
([Error translation](http.md#error-translation)). That is how login, change
password, refresh and verify answer every failure with one code while the log
tells the failures apart ([reasons](../logging.md#outcomes)).

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
  against a dummy hash, so a missing account costs a bcrypt comparison like a
  wrong password does
  ([decision 0045](../development/decisions/0045-login-dummy-hash.md)).
- Every rejection is `401 invalid_credentials`, each with its own reason.
  `ErrPasswordChanged` is the one the use cases return themselves, when the
  password is replaced between the check and their write.

Each use case builds its `Checker` in `New`. `internal/bootstrap` creates one
dummy hash at startup, with the configured hasher, and gives it to both.

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

Step 3 is atomic, so a refresh token never exists without its session. It
first locks the password row, only while its hash is still the one step 1
checked. Step 1 takes hundreds of milliseconds, and a password change that
committed in between would otherwise miss the session about to be saved, which
would stay refreshable for good with a password that no longer works. With the
lock, a change that comes later waits for this transaction, and its revocation
reaches the session; a change that came first leaves nothing to lock, and the
login rolls back and answers `401` with the reason `password changed`, as it
would have a moment later
([decision 0055](../development/decisions/0055-login-locks-the-password-it-checked.md)).

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
covered by [login's lock](#login): no session opened with the old password
outlives the change.

## Refresh

`internal/usecase/refresh` exchanges a refresh token for a new pair, and
carries reuse detection.

1. Parse the token, hash it, `FindByHash`. Malformed or unknown:
   `invalid token`.
2. Load its session. A token without one is a bug: `500`.
3. Session revoked: `session revoked`.
4. `token.Use(now)`. Expired: `token expired`. Already used: reuse detection.
   Otherwise, rotate.

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
requests with the same token can both pass step 4, and exactly one rotates. The
other is treated as reuse
([decision 0050](../development/decisions/0050-refresh-token-use-is-settled-at-write.md)).

### Reuse detection

A refresh token is used once, so a second presentation means two parties hold
it: the client, and someone who copied it. The service cannot tell which one
is calling, so it revokes the whole session, which locks out both: the attacker
for good, the user until they log in again. Serving the request instead would
let an attacker who replayed a token keep a rotating credential indefinitely.

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
`infra/jwt.AccessTokenService.Validate`: EdDSA, a published `kid`, the
signature and `exp`, then `sub` as the user ID. A malformed or forged token, or
one signed by an unknown key, is `invalid token`; an expired one is
`token expired`; both answer `401 invalid_token`, and anything else `500`.
`ExpiresIn` is never negative, because the clock is read before the token is
checked.

It reads no database: it runs the code a local verifier is pointed to as a
reference ([What to verify](../api/token-verification.md#what-to-verify)), so a
token passes here exactly when it would pass a correct verifier elsewhere, and a
service can switch between the two. An access token names a user, not a
session, so the tokens of a revoked session pass until their `exp`. It is never
rate limited ([why](http.md#rate-limiting)).
