# Signing keys are shared through the database, and the clock decides which one signs

- **Status:** Accepted
- **Date:** 2026-10-06
- **Related:** [Signing keys](../../architecture/tokens.md#signing-keys), [Guarded writes](../../architecture/persistence/repositories.md#guarded-writes), [0050](0050-refresh-token-use-is-settled-at-write.md)

## Invariant

Every signing key is a row of `signing_keys`, added only through
`SigningKeyStore.Add`, which applies only while its generation is free, after
which the keyring lists the keys again. Every instance signs with the key whose
`active_at` is the latest not after its own clock, or, while none is active
yet, the first to activate. A key other than the first is added a day before
its `active_at`, and a key is deleted only once a later key has been active for
25 hours, longer than any access token lives.

## Context

The keyring held one key, generated at startup and kept in process memory. Each
instance signed with its own key and published only that one, so a token from
one instance failed against another's JWKS, and a restart invalidated every
outstanding access token. Rotation, every 7 days, replaced the key in one step,
which broke tokens in two ways:

- **the old key left the JWKS at once**, so the tokens it had signed stopped
  verifying before they expired;
- **the new key was published only when it started signing**, so a verifier
  holding a cached JWKS rejected fresh tokens until it fetched the document
  again.

Sharing the keys needs a store every instance reaches, and a way for the
instances to agree on which key signs without talking to each other. Verifiers
cache the JWKS for as long as they choose: the response has no cache headers,
and nothing tells the service when they fetch it.

## Decision

The keys are rows of `signing_keys`, in the PostgreSQL database every instance
already shares, behind `port.SigningKeyStore`: `postgres.SigningKeyRepo`
implements it and the keyring consumes it, so the repositories stay the only
code that writes SQL. The 32-byte seed is stored as it is.

**The clock decides which key signs.** Each key has a `generation` and an
`active_at`, and signs from its `active_at` until the next key's. Every
instance reads the same rows, so every instance switches at the same moment,
give or take clock skew, with no write and no message to tell it to. A key
stops signing only once a successor exists, so a rotation that fails leaves
the current key signing, never none.

**The windows cover the caches.** The next key is added a day before its
`active_at`. Every instance has loaded it for most of that day before it
signs, and so has a verifier that fetches the JWKS at least every 23 hours. A
key is deleted once a later key has been active for 25 hours: an hour past the
longest `ACCESS_TOKEN_TTL` that `infra/jwt` accepts, whatever TTL each instance
runs with, so no valid token outlives its key. Deleting the key, rather than
keeping it, limits what a leaked database or backup gives out to the keys of
the last 9 days or so. The windows are constants in `internal/bootstrap`. How
a key moves through them is in [Signing keys](../../architecture/tokens.md#signing-keys).

**The first key signs at once**, since nothing else can. An instance whose
clock lags the peer that added it finds no key active yet, and then signs with
the first to activate, rather than failing logins for the seconds of skew.

**The primary key settles a race.** Two instances can decide to add the same
generation at once. The insert is a guarded write,
`ON CONFLICT (generation) DO NOTHING`, of the shape
[Guarded writes](../../architecture/persistence/repositories.md#guarded-writes)
describes. `Add` reports nothing either way, because the keyring lists the
keys again after every `Add`: both instances sign with the row that was kept,
never with the key one of them generated.

`Keyring.Sync` does all of it: list, add the next key when it is due, list
again, publish the keys that are not retired, then delete the retired ones.
The `jwt_keyring_rotation` task runs it every 10 minutes, and startup runs it
without the delete. A failed list or add leaves the published keys as they
were; the delete comes after they are published, so a failed one stalls
neither rotation nor startup.

## Alternatives considered

- **An optional secret to encrypt the seeds**: a new variable has to be
  optional to keep v1 compatible, so the default deployment would stay
  unencrypted either way, for crypto code and key management on top. Every key
  is deleted within about 9 days, so encryption can come later without
  migrating any data.
- **A file or a volume**: not shared between hosts.
- **The PostgreSQL adapter importing `ed25519`, or SQL in `ed25519`**: the
  first import between adapters, or SQL outside the repositories.
- **A `sign_until` and a `publish_until` stored when a key is added**: a
  rotation that failed would leave no key signing once `sign_until` passed.
- **A "current" flag that rotation flips**: every instance would have to
  reload at the same instant, or sign for a while with a key the others no
  longer use.
- **An advisory lock around the check and the insert**: a second concurrency
  shape, held across round trips, for what the primary key already settles.
- **`INSERT … WHERE NOT EXISTS`**: not atomic under `READ COMMITTED`.
- **One key per instance, all published**: the JWKS grows with the number of
  instances, and an instance that goes away leaves its key behind.
- **The windows as environment variables**: additive, so they can come later,
  and nothing needs them now.
- **A retire window derived from the configured `ACCESS_TOKEN_TTL`**: instances,
  and restarts, can run with different TTLs.
- **A lead of an hour**: shorter for no gain, and some verifier caches hold the
  JWKS for hours.

## Consequences

- Whoever reads the database, or a backup of it, can sign access tokens that
  every verifier accepts while that key is published
  ([Limitations](../../limitations.md#signing-keys-are-stored-in-the-database)).
- The instances' clocks have to agree within an hour, the margin the windows
  leave.
- Every instance syncs every 10 minutes: two or three queries, and a log line.
- Retired keys are deleted only when the service's database role can `DELETE`
  from `signing_keys`. Without it they stay, and every sync logs a failure.
- The first start after upgrading adds a first key, so it invalidates
  outstanding access tokens once, as every restart did before.
- Withdrawing a leaked key has no procedure: deleting its row leaves the
  instances disagreeing for up to one sync interval.
