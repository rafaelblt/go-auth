# Tokens

`internal/infra/jwt`, `internal/infra/jwt/ed25519`

The two credentials the service hands out, the keys that sign one of them, and
the security properties that follow.

## Token model

Two tokens, with opposite trade-offs: a token that can be verified without
calling this service cannot be revoked, and a token that can be revoked has to
be checked here.

| | Access token | Refresh token |
|---|---|---|
| Format | JWT, EdDSA (Ed25519) | 32 random bytes, base64url |
| Verified by | any service, with the JWKS, or this service on request | this service only |
| Lifetime | `ACCESS_TOKEN_TTL`, default 30m | `REFRESH_TOKEN_TTL`, default 7 days |
| Stored | no | as its SHA-256 hash |
| Use | until it expires | once: every refresh replaces it |
| Revocable | no | yes, with its session |

The access token is what services check on every request, without a call to
`go-auth`; the refresh token is where control lies. So `ACCESS_TOKEN_TTL` is
how long a revoked session can still present a valid access token
([Limitations](../limitations.md#access-tokens-cannot-be-revoked)).

The refresh token is random bytes rather than a JWT: it is looked up on every
use, so it needs no content, only to be impossible to guess, and a JWT would
invite clients to read claims from it. Only its SHA-256 is stored, so a leaked
database gives out digests rather than working tokens. A slow hash would add
nothing: there is nothing to guess in 32 random bytes, and bcrypt would cost
every refresh hundreds of milliseconds.

## Access tokens

`jwt.AccessTokenService` issues them with two claims, `sub` and `exp`
([Token format](../api/token-verification.md#token-format)), and a lifetime
that it accepts only between 1 minute and 24 hours (`MaxExpiration`).
`ed25519.Signer` signs them with the key the keyring says signs now, putting its
ID in the `kid` header. Validation pins `EdDSA`, picks the key by `kid` among the
published ones and requires `exp`, by the injected clock; the verify endpoint
runs the same code.

## Signing keys

Ed25519, because verifiers outside this service hold the keys: with HMAC,
every verifier would hold the signing secret and could mint tokens. Ed25519
keys and signatures are much smaller than RSA's, and JWKS supports them
(`OKP`) as standard.

The keys are rows of `signing_keys`, in the database every instance already
shares, read and written through `port.SigningKeyStore`. So they survive a
restart, and every instance signs and publishes the same ones: with keys held in
memory, a token would fail against another instance, or after a restart. A
key's ID is its RFC 7638 thumbprint, derived from the key. Each key goes
through three windows:

- it is added, and published, a day before its `active_at`, so by the time it
  signs, every instance has loaded it, and so has any verifier that fetches the
  JWKS at least every 23 hours. The first key is the exception: it signs at
  once, since nothing else can;
- it signs from its `active_at` until the next key's, 7 days later;
- it stays published for 25 hours more, an hour longer than the longest access
  token `infra/jwt` accepts, whatever TTL each instance runs with, and is then
  deleted, so a copy of the database gives out only the keys of the last 9 days
  or so.

A key is therefore published for about 9 days, and the JWKS holds one or two
keys. The windows are constants in `internal/bootstrap/infra.go`.

**The clock decides which key signs**: each instance signs with the key whose
`active_at` is the latest not after its own clock. Every instance reads the
same rows, so all of them switch at the same moment, give or take their clock
skew, with no write and no message. A key stops signing only once a successor
exists, so a rotation that fails leaves the current key signing, never none. An
instance whose clock lags the one that added the first key finds no key active
yet, and signs with the first to activate rather than failing logins.

`Keyring.Sync` runs every 10 minutes on each instance
([Background tasks](startup.md#background-tasks)): it lists the keys, adds the
next one when it is due, lists them again, publishes those not retired, and then
deletes the retired ones. Startup loads the keys the same way, without the
delete. When two instances add the same generation, the primary key keeps the
first ([Guarded writes](persistence.md#guarded-writes)), and both sign with the
row read back, never with a key one of them only generated. Signing reads the
keys through an `atomic.Pointer`, and only a sync takes a lock.

A sync that fails changes nothing: the instance keeps the keys it last loaded,
and switches to one it already holds when that key's time comes, but does not
see keys other instances add. If every instance keeps failing, the current key
signs past its 7 days. The delete comes last, so one that fails, for instance
because the database role cannot `DELETE`, leaves the retired keys in the table
and nothing else. A key cannot be withdrawn while the service runs, since
deleting its row leaves the instances disagreeing until each has synced:
withdrawing one means
[replacing every key](../configuration.md#replacing-the-signing-keys) with the
instances stopped.

Considered and rejected:

- **A "current" flag that rotation flips**: every instance would have to reload
  at the same instant, or sign for a while with a key the others no longer use.
- **A `sign_until` and a `publish_until` stored with each key**: a rotation that
  failed would leave no key signing once `sign_until` passed.
- **One key per instance, all published**: the JWKS grows with the number of
  instances, and an instance that goes away leaves its key behind.
- **A retire window derived from `ACCESS_TOKEN_TTL`**: instances, and restarts,
  can run with different TTLs.

### Sealed seeds

When [`SIGNING_KEY_ENCRYPTION_KEY`](../configuration.md#signing_key_encryption_key)
is set, the keyring seals each seed it adds with AES-256-GCM (a random 12-byte
nonce and a 16-byte tag, with the key's generation as additional data, so a
sealed seed opens only in its own row), and opens the stored ones as it reads
them. The repository stores whatever bytes it is given.

The length tells the two forms apart: 32 bytes is a plaintext seed and 60 a
sealed one, so no column records it. A plaintext seed loads whether or not the
key is set, which is what lets a running deployment set it with nothing else to
do: the keys stored before age out within about 9 days. Sealing protects a copy
of the database, not the database itself: a plaintext row written into the
table still loads, by whoever could just as well replace a password hash.

## Security properties

| Property | Holds |
|---|---|
| Passwords stored as bcrypt hashes, at a configurable cost | yes |
| Refresh tokens stored only as their SHA-256 | yes |
| Refresh tokens single use, replaced on every refresh | yes |
| Reusing a refresh token revokes its session | yes |
| Changing the password revokes every session of the user | yes |
| Login and change password answer every credential failure alike | yes ([why](../api/errors.md#deliberately-vague-errors)) |
| A missing account takes as long as a wrong password | yes, until `BCRYPT_COST` changes ([why](../limitations.md#login-timing-after-a-bcrypt_cost-change)) |
| Whether a username exists is secret | **no**: registration says it ([why](../limitations.md#registration-reveals-whether-a-username-exists)) |
| Refresh failures look alike in the response | yes |
| Access tokens can be revoked | **no** ([why](../limitations.md#access-tokens-cannot-be-revoked)) |
| Signing keys survive a restart, and rotate without breaking issued tokens | yes |
| Signing keys encrypted at rest | only with [`SIGNING_KEY_ENCRYPTION_KEY`](../configuration.md#signing_key_encryption_key) |
| Rate limiting | only with [`RATE_LIMIT`](../configuration.md#rate_limit) |
| Maximum session age | **no** ([why](../limitations.md#sessions-have-no-maximum-age)) |
