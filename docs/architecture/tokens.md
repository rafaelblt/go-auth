# Tokens

`internal/infra/jwt`, `internal/infra/jwt/ed25519`

The code that issues and checks access tokens, and the keys that sign them. What
the two tokens are and why: [The two tokens](../system.md#the-two-tokens). The
refresh token has no code here: it is a domain type
([`session`](domain.md#session)).

## Access tokens

`internal/infra/jwt` implements `port.AccessTokenIssuer` and
`port.AccessTokenValidator`, which are what the use cases see.

| File | Holds |
|---|---|
| `access_token_service.go` | `AccessTokenService`: `Issue`, `Validate`; `CheckExpiration`, `MaxExpiration` |
| `errors.go` | `ErrTokenInvalid`, `ErrTokenExpired`, which the signer returns |

**Rules the code keeps**

- `Issue` sets two claims, `sub` and `exp`
  ([Token format](../access-tokens.md#token-format)), with a lifetime
  `CheckExpiration` accepts only between 1 minute and `MaxExpiration`
  ([`ACCESS_TOKEN_TTL`](../configuration.md#access_token_ttl)), and
  `NewAccessTokenService` calls. `MaxExpiration` also sets how long a retired
  key stays published ([Signing keys](#signing-keys)).
- Validation implements the
  [checks of verify](../access-tokens.md#letting-go-auth-verify).
  `ed25519.Signer.Parse` makes all of them but one: it pins `EdDSA`, picks the
  key by `kid` among the published ones, requires `exp`, checks `nbf` when
  there is one, by the injected clock, and maps each `golang-jwt` failure to
  `ErrTokenInvalid` or `ErrTokenExpired`. `AccessTokenService.Validate` then
  parses `sub` as a `user.ID`, and maps everything to `session.ErrTokenInvalid`
  or `session.ErrTokenExpired`. Any other error is unexpected.

**When you change it:** a change to what is accepted changes
[Letting go-auth verify](../access-tokens.md#letting-go-auth-verify), and, if a
verifier elsewhere should follow, [What to verify](../access-tokens.md#what-to-verify).
A `golang-jwt` failure left out of the mapping becomes an
`internal_server_error`.

## Signing keys

`internal/infra/jwt/ed25519` holds the keys, the keyring that loads, rotates and
serves them, and the JWT signer. Implements [Signing keys](../access-tokens.md#signing-keys).

| File | Holds |
|---|---|
| `keyring.go` | `Keyring`: `NewKeyring`, `Sync`, `SigningKey`, `PublicKeyByID`, `PublicKeys` |
| `key.go` | the `key` type, seed generation, the RFC 7638 thumbprint that is a key's ID |
| `signer.go` | `Signer`: `Sign`, `Parse` |
| `seed_cipher.go` | sealing and opening seeds ([Sealed seeds](#sealed-seeds)); `CheckEncryptionKey`, none or 32 bytes |
| `errors.go`, `consts.go` | the parse errors; the algorithm, curve and key type names |

**Rules the code keeps**

- The keys are rows of `signing_keys`, read and written through
  `port.SigningKeyStore`. A key's ID is derived from the key, so it is not
  stored.
- `SigningKey` returns the key whose `active_at` is the latest not after the
  instance's clock. An instance whose clock lags the one that added the first
  key finds no key active yet, and signs with the first to activate rather than
  failing logins.
- **`Sync`** lists the keys, adds the next one when it is due, lists them again,
  publishes those not retired, and then deletes the retired ones. Startup loads
  the keys the same way, without the delete. When two instances add the same
  generation, the primary key keeps the first
  ([Guarded writes](persistence.md#guarded-writes)), and both sign with the row
  read back, never with a key one of them only generated.
- Signing reads the keys through an `atomic.Pointer`, and only `Sync` takes a
  lock.
- A failed `Sync` changes nothing: the instance keeps the keys it last loaded,
  and switches to one it already holds when that key's time comes, but does not
  see keys other instances add. The delete comes last, so one that fails, for
  instance because the database role cannot `DELETE`, leaves the retired keys
  in the table and nothing else.
- The windows are `signingKeyPublishBefore`, `signingKeyRotationInterval` and
  `signingKeyPublishAfter`, the last derived from `jwt.MaxExpiration`, in
  `internal/bootstrap/infra.go`, and reach the keyring through its `Config`. `Sync` runs as a
  [background task](startup.md#background-tasks).

**When you change it:** a window changes
[Signing keys](../access-tokens.md#signing-keys) and the caching advice in
[Fetching and caching keys](../access-tokens.md#fetching-and-caching-keys).

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

When given an encryption key, the keyring seals each seed it adds with
AES-256-GCM: a random 12-byte nonce and a 16-byte tag, with the key's generation
as additional data, so a sealed seed opens only in its own row. It opens the
stored ones as it reads them. The repository stores whatever bytes it is given.

The length tells the two forms apart: 32 bytes is a plaintext seed and 60 a
sealed one, so no column records it. A plaintext seed loads whether or not the
key is set ([why](../access-tokens.md#sealed-keys)).
