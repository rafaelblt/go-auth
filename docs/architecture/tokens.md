# Tokens

The two credentials the service hands out, how the access token is signed, and
the security properties that follow from both.

## Token model

Two tokens, with opposite trade-offs. A token that can be verified without
calling this service cannot be revoked, and a token that can be revoked has to
be checked here. No single token has both properties.

| | Access token | Refresh token |
|---|---|---|
| Format | JWT, EdDSA (Ed25519) | 32 random bytes, base64url |
| State | stateless | stored in the database |
| Verified by | any service, locally, with the JWKS | this service only |
| Lifetime | short (default 30m) | long (default 7d) |
| Revocable | no | yes |
| Stored as | not stored | SHA-256 hash |
| Reusable | yes, until `exp` | no, single use |

The access token is what your services check on every request, and it cannot
be revoked, so checking it needs no call to `go-auth`. The refresh token is
where control lies: it is stored, used once, and revoking its session stops
the client from getting new access tokens.

`ACCESS_TOKEN_TTL` is the gap between the two: how long a revoked session can
still present a valid access token. That is why the default is 30 minutes,
not a day.

**The refresh token is random bytes, not a JWT.** It is looked up in the
database on every use, so it needs no content of its own. Its only job is to
be impossible to guess. A JWT would be larger, and would invite clients to
read claims from it.

**Only its SHA-256 hash is stored**, so a leaked database gives out digests,
not working tokens. A slow hash like bcrypt is unnecessary: the token is 32
bytes of uniform randomness, with nothing to guess, and bcrypt would add
hundreds of milliseconds to every refresh.

The value objects behind both are in
[Session and refresh tokens](domain/session.md). How they are issued and
rotated is in [Login](usecases/login.md) and [Refresh](usecases/refresh.md).

## Signing keys

`internal/infra/jwt/ed25519`

Access tokens are signed with Ed25519. They are verified outside this
service, so the algorithm decides what verifiers must hold. With HMAC, every
verifier would hold the signing secret, and any of them could mint tokens.
Ed25519 keys and signatures are much smaller than RSA's, signing is faster,
and JWKS supports its keys (`OKP`) as standard.

The `Keyring` holds every published key, loaded from the `signing_keys`
table through `port.SigningKeyStore`, so the keys survive a restart and every
instance signs with the same ones. A key's ID is its RFC 7638 thumbprint,
derived from the key itself, so it needs no storage of its own. The signer
puts the ID in the JWT's `kid` header, and the keyring publishes the matching
public keys at `/.well-known/jwks.json`, so a verifier picks the right key
without trying each one.

Each key goes through three windows:

- it is added, and published, a day before its `active_at`. The first key is
  the exception: it signs at once;
- it signs from its `active_at` until the next key's, 7 days later;
- it stays published for 25 hours more, longer than any access token lives,
  and is then deleted.

Every 10 minutes each instance reads the table again, adds the next key when
it is due, and deletes the retired ones
([Startup](startup.md#background-tasks)). Each instance's own clock decides
which key signs, so the switch needs no coordination: by then every instance
has had the next key for most of a day
([why](../development/decisions/0053-signing-keys-are-shared-through-the-database.md)). When two instances add the same
generation, its primary key keeps the first, and both read it back
([guarded writes](persistence/repositories.md#guarded-writes)).

Reads go through an `atomic.Pointer`, and only a sync takes a lock, so
signing never waits for it. The keys are stored unencrypted
([Limitations](../limitations.md#signing-keys-are-stored-in-the-database)).

## Security properties

| Property | Holds |
|---|---|
| Passwords stored hashed (bcrypt, configurable cost) | yes |
| Refresh tokens never stored, only their SHA-256 | yes |
| Refresh tokens single use, replaced on every refresh | yes |
| Reusing a refresh token revokes its session | yes |
| Login errors do not reveal whether an account exists | yes |
| Login timing does not reveal whether an account exists | yes, unless `BCRYPT_COST` changed ([why](../limitations.md#login-timing-after-a-bcrypt_cost-change)) |
| Refresh errors do not reveal a token's state | yes |
| Access tokens can be revoked | **no** ([why](../api/token-verification.md#revocation-is-not-immediate)) |
| Signing key survives a restart | yes |
| Key rotation keeps issued tokens valid | yes |
| Signing keys encrypted at rest | **no** ([why](../limitations.md#signing-keys-are-stored-in-the-database)) |
| Rate limiting | only when [`RATE_LIMIT`](../configuration.md#rate_limit) is set; otherwise use a reverse proxy ([why](../limitations.md#run-it-behind-a-reverse-proxy)) |
| Maximum session age | **no** ([why](../limitations.md#sessions-have-no-maximum-age)) |
