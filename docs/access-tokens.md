# Access tokens

What the user's requests carry to your services. An access token holds
everything needed to check it, so your services verify it with the public keys
`go-auth` publishes, without calling it: `go-auth` is involved in logins and
refreshes, not in the requests your application serves. A service that would
rather not verify tokens itself can [let `go-auth` do it](#letting-go-auth-verify).
How they compare with refresh tokens: [The two tokens](system.md#the-two-tokens).
What every endpoint shares: [API conventions](api/conventions.md).

## Token format

```
Header    { "alg": "EdDSA", "typ": "JWT", "kid": "NzbLsXh8uDCcd-6MNwXF4W..." }
Payload   { "sub": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "exp": 1789567800 }
Signature Ed25519 over base64url(header) + "." + base64url(payload)
```

| Claim | Meaning |
|---|---|
| `sub` | The user's ID, a UUID |
| `exp` | When the token expires, as a Unix timestamp: [`ACCESS_TOKEN_TTL`](configuration.md#access_token_ttl) after it was issued |

That is the whole payload: no `iss`, `aud`, `iat` or `jti`, no username, no
roles. If your application needs more than the user ID, look it up by `sub`, or
ask `go-auth` with [me](#looking-up-the-user). The `kid` header names the
[key](#signing-keys) that signed the token.

**Why:** a deployment serves one application, so there is no issuer or audience
to tell apart, and a claim that verifiers are told to ignore only invites
careless verification.

## What to verify

A correct verifier checks all of these:

1. **The algorithm is `EdDSA`.** Pin it. Trusting the algorithm the token names
   is the classic JWT vulnerability: a token with `"alg": "none"`, or with an
   HMAC algorithm keyed with the public key, must be rejected.
2. **The signature is valid** under the published key whose `kid` matches the
   token's header.
3. **`exp` is present and in the future.** A token without `exp` must be
   rejected, not treated as never expiring.
4. **`sub` is a UUID.**

Mainstream JWT libraries do 2 and 3; 1 usually needs an explicit option, and 4
is up to you. `go-auth` checks its own tokens the same way, in code that can
serve as a reference: `internal/infra/jwt/ed25519/signer.go` makes checks 1 to
3, and `internal/infra/jwt/access_token_service.go` check 4.

## Fetching and caching keys

The keys are the [JWKS](#the-jwks); `x` is the raw 32-byte Ed25519 public key,
in base64url without padding.

Cache the document: fetching it on every request defeats the purpose. It has no
cache headers, so choose the lifetime yourself: cache it for a few minutes,
fetch it again when a token has a `kid` you do not know, and limit how often
that refetch can happen, so tokens with random `kid`s cannot force a flood of
requests.

A key is in the document a day before it starts signing, and stays after it
stops for longer than any access token lives ([Signing keys](#signing-keys)).
So a verifier that fetches the document at least every 23 hours never meets a
valid token whose `kid` it lacks, and no valid token loses its key.

## Examples

### Go

```go
import (
    "crypto/ed25519"
    "errors"

    "github.com/golang-jwt/jwt/v5"
)

// keys maps kid to public key, filled from the JWKS document.
func verify(raw string, keys map[string]ed25519.PublicKey) (string, error) {
    token, err := jwt.ParseWithClaims(raw, &jwt.RegisteredClaims{},
        func(t *jwt.Token) (any, error) {
            kid, _ := t.Header["kid"].(string)
            key, ok := keys[kid]
            if !ok {
                return nil, errors.New("unknown kid")
            }
            return key, nil
        },
        jwt.WithValidMethods([]string{"EdDSA"}), // pin the algorithm
        jwt.WithExpirationRequired(),            // reject tokens without exp
    )
    if err != nil {
        return "", err
    }
    return token.Claims.(*jwt.RegisteredClaims).Subject, nil
}
```

### Node.js

```js
import { createRemoteJWKSet, jwtVerify } from 'jose'

const jwks = createRemoteJWKSet(
  new URL('http://localhost:8080/.well-known/jwks.json')
)

export async function verify(token) {
  const { payload } = await jwtVerify(token, jwks, {
    algorithms: ['EdDSA'],        // pin the algorithm
    requiredClaims: ['exp'],
  })
  return payload.sub
}
```

`jose` fetches, caches and selects keys by `kid` for you.

### Python

```python
import jwt  # PyJWT
from jwt import PyJWKClient

jwks = PyJWKClient("http://localhost:8080/.well-known/jwks.json")

def verify(token: str) -> str:
    key = jwks.get_signing_key_from_jwt(token).key
    claims = jwt.decode(
        token,
        key,
        algorithms=["EdDSA"],                       # pin the algorithm
        options={"require": ["exp", "sub"], "verify_exp": True},
    )
    return claims["sub"]
```

PyJWT needs its crypto extra for Ed25519: `pip install "pyjwt[crypto]"`.

## Letting go-auth verify

`POST /v1/auth/verify` checks an access token for a service that would rather
not verify it itself, because its language lacks a JWT library you trust, or to
avoid fetching and caching the keys. It answers with the user the token was
issued to.

A token passes when it passes the [checks above](#what-to-verify) and, when it
carries an `nbf`, that time has come. It reads no database, so it accepts
exactly the tokens a correct local verifier accepts, and a service can switch
between the two. It is never rate limited.

**Request**

```json
{ "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..." }
```

**Response: `200 OK`**

```json
{
  "user_id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37",
  "expires_at": "2026-09-08T12:30:00Z",
  "expires_in": 1342
}
```

| Field | Meaning |
|---|---|
| `user_id` | The user the token was issued to: its `sub` |
| `expires_at` | When the token expires: its `exp` |
| `expires_in` | Whole seconds from the moment the service checked the token to its expiry, rounded down |

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_token` | Any failure, a malformed or expired token included: turn away the request that carried it |

The log tells an expired token from the others ([reasons](logging.md#outcomes)).

It costs what verifying locally saves: a round trip to `go-auth` on every
request your service serves, and those requests fail while `go-auth` cannot be
reached. An answer holds until `expires_at`
([Access tokens cannot be revoked](#access-tokens-cannot-be-revoked)), so you
can keep it for that token until then and skip most of the round trips.

**Why no rate limit:** it reads no database, no number of tries finds a
signature that passes, and the services that delegate checking call it on every
request they serve, usually from one address.

## Looking up the user

`POST /v1/auth/me` returns the user an access token was issued to, for a client
that holds the token and needs more than the user ID it carries, such as the
username to show. It checks the token [as verify does](#letting-go-auth-verify),
then reads the user. It reads nothing about the token's session.

**Request**

```json
{ "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..." }
```

**Response: `200 OK`**, the same `user` as [registration](accounts.md#registration)
returns:

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_token` | Whatever verify rejects, and a token whose user is no longer stored: refresh, and send the user to log in if that fails too |

Nothing in the service deletes a user, so a missing one takes a change made
outside it, such as a row deleted by hand or a database restored from an older
backup.

Unlike verify, it is rate limited. **Why:** every token that passes costs a
database read, and it serves the clients that hold the tokens, which ask when
they need to, not on every request a service serves.

## The JWKS

`GET /.well-known/jwks.json` returns the public keys that verify access tokens,
as a [JWK Set](https://www.rfc-editor.org/rfc/rfc7517). It holds no secrets and
needs no authentication.

```json
{
  "keys": [
    {
      "kty": "OKP",
      "crv": "Ed25519",
      "x": "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
      "use": "sig",
      "alg": "EdDSA",
      "kid": "NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"
    }
  ]
}
```

| Field | Meaning |
|---|---|
| `kty`, `crv` | Always `OKP` and `Ed25519` |
| `x` | The public key, base64url without padding |
| `use`, `alg` | Always `sig` and `EdDSA` |
| `kid` | Key ID: the RFC 7638 thumbprint of the key, and the `kid` in the header of the tokens it signed |

It usually holds one or two keys ([Signing keys](#signing-keys)), and has no
cache headers ([caching](#fetching-and-caching-keys)).

## Signing keys

The keys are rows in the database every instance shares, so they survive a
restart, and every instance signs and publishes the same ones. Each key goes
through three windows:

- it is added, and published, **a day** before it starts signing, so by then
  every instance has loaded it, and so has a verifier that fetches the JWKS
  [often enough](#fetching-and-caching-keys). The first key is the exception:
  it signs at once, since nothing else can;
- it signs for **7 days**, until the next key starts;
- it stays published for **25 hours** more, an hour longer than the longest
  access token the service can issue, and is then deleted, so a copy of the
  database holds only the keys of the last 9 days or so.

So a key is published for about **9 days**, and the JWKS holds one or two keys.
Each instance re-reads the keys every few minutes
([interval](architecture/startup.md#background-tasks)), and goes by its own
clock to decide which key signs: they all switch at the same moment, give or
take their clock skew, with no message between them. The extra hour of the last
window absorbs that skew: with more, a token one instance signed just before it
switched could lose its key at another while still valid. A key stops signing
only once its successor exists, so a rotation that fails leaves the current key
signing, never none.

**Why Ed25519:** the services that check the tokens hold only public keys; with
HMAC every one of them would hold the signing secret and could mint tokens.
Ed25519 keys and signatures are small, and the JWKS format supports them as
standard.

### Sealed keys

When [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key)
is set, every key added is stored encrypted, and opens only in its own row. A
key stored before it was set still loads, so setting it needs nothing else: the
old keys age out as their windows end. Sealing protects a copy of the database,
not the database itself: whoever can write the table can store a key in the
clear, as they could replace a password hash.

## Limitations

### Access tokens cannot be revoked

Revoking a session ([how](sessions.md#life-of-a-session)) stops its refresh
tokens at once, but the access tokens already issued stay valid until they
expire: they hold everything needed to check them, not even their session.
`ACCESS_TOKEN_TTL` is therefore how long a revoked session keeps working.
[Verify](#letting-go-auth-verify) and [me](#looking-up-the-user) do not change
that: they make the checks a local verifier makes, and me then reads only the
user. An application that needs revocation to take effect at once cannot get it
from access tokens.

### Signing keys are stored in the database

- **They are unencrypted unless `SIGNING_KEY_ENCRYPTION_KEY` is set**
  ([Sealed keys](#sealed-keys)). Without it, whoever reads the database, or a
  backup of it, can sign access tokens that every verifier accepts while that
  key is published. Protect the database and its backups like the keys they
  hold.
- **No key can be withdrawn while the service runs, and the encryption key
  cannot be changed in place**: either means
  [replacing every signing key](configuration.md#replacing-the-signing-keys),
  which invalidates the access tokens already issued.
- **Instances go by their own clocks** to decide which key signs and when an
  old one is dropped, and the windows absorb only an hour of skew
  ([Signing keys](#signing-keys)): keep the clocks synchronised.
