# Signing keys are shared through the database, and the clock decides which one signs

- **Status:** Accepted

## Invariant

Every signing key is a row of `signing_keys`, added only through
`SigningKeyStore.Add`, which applies only while its generation is free, after
which the keyring lists the keys again. Every instance signs with the key whose
`active_at` is the latest not after its own clock, or, while none is active
yet, the first to activate. A key other than the first is added a day before
its `active_at`, and a key is deleted only once a later key has been active for
25 hours, longer than any access token lives.

## Context

The keyring used to hold one key, generated at startup in process memory. Each
instance signed with its own key and published only that one, so a token from
one instance failed against another's JWKS, and a restart invalidated every
access token. Rotation replaced the key in one step: the old key left the JWKS
at once, so the tokens it had signed stopped verifying before they expired, and
the new one was published only when it started signing, so a verifier holding a
cached JWKS rejected fresh tokens.

Sharing the keys needs a store every instance reaches, and a way for instances
to agree on which key signs without talking to each other. Verifiers cache the
JWKS for as long as they choose, and nothing tells the service when they fetch
it.

## Decision

The keys are rows in the PostgreSQL database every instance already shares,
behind `port.SigningKeyStore`, so the repositories stay the only code that
writes SQL and no adapter imports another. How a key moves through its windows
is in [Signing keys](../../architecture/tokens.md#signing-keys).

**The clock decides which key signs.** Each key signs from its `active_at`
until the next key's. Every instance reads the same rows, so every instance
switches at the same moment, give or take clock skew, with no write and no
message to tell it to. A key stops signing only once a successor exists, so a
rotation that fails leaves the current key signing, never none.

**The windows cover the caches.** The next key is added a day before it signs,
rather than an hour, because some verifiers cache the JWKS for hours: by the
time it signs, every instance has loaded it, and so has any verifier that
fetches the JWKS at least every 23 hours. A key is deleted 25 hours after its
successor starts, an hour past the longest access token `infra/jwt` accepts,
whatever TTL each instance runs with. Deleting rather than keeping it limits
what a leaked database or backup gives out to the keys of the last 9 days or so.

**The first key signs at once**, since nothing else can. An instance whose
clock lags the one that added it finds no key active yet, and signs with the
first to activate rather than failing logins for the seconds of skew.

**The primary key settles a race.** Two instances can add the same generation
at once. The insert is a
[guarded write](../../architecture/persistence.md#guarded-writes) that reports
nothing, because the keyring lists the keys again after every `Add`: both
instances sign with the row that was kept, never with the key one of them
generated.

## Alternatives

- **A "current" flag that rotation flips**: every instance would have to reload
  at the same instant, or sign for a while with a key the others no longer use.
- **A `sign_until` and a `publish_until` stored when a key is added**: a
  rotation that failed would leave no key signing once `sign_until` passed.
- **One key per instance, all published**: the JWKS grows with the number of
  instances, and an instance that goes away leaves its key behind.
- **A retire window derived from the configured `ACCESS_TOKEN_TTL`**: instances,
  and restarts, can run with different TTLs.

## Consequences

- Whoever reads the database, or a backup of it, can sign access tokens while
  that key is published, unless the seeds are
  [sealed](../../architecture/tokens.md#sealed-seeds).
- The instances' clocks have to agree within an hour.
- Every instance syncs every 10 minutes: two or three queries, and a log line.
- Retired keys are deleted only when the database role can `DELETE` from
  `signing_keys`; without it they stay, and every sync logs a failure.
- A key cannot be withdrawn while the service runs: deleting rows then leaves
  the instances disagreeing for up to one sync. The way to withdraw one is to
  [replace every key](../../configuration.md#replacing-the-signing-keys) with
  the instances stopped.
