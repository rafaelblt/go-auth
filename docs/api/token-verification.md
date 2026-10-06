# Verifying access tokens

Access tokens are JWTs signed with Ed25519, and they hold everything needed to
check them. Your services verify them with the public key published at
`/.well-known/jwks.json`, without calling `go-auth` on every request.

That is the point of the design: `go-auth` is involved in login and refresh,
which happen rarely, and not in the requests your application serves.

## Token format

```
Header    { "alg": "EdDSA", "typ": "JWT", "kid": "NzbLsXh8uDCcd-6MNwXF4W..." }
Payload   { "sub": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "exp": 1789567800 }
Signature Ed25519 over base64url(header) + "." + base64url(payload)
```

| Claim | Meaning |
|---|---|
| `sub` | The user's ID, a UUID string |
| `exp` | When the token expires, as a Unix timestamp |

That is the whole payload: no `iss`, `aud`, `iat` or `jti`, no username, no
roles or scopes. A deployment serves one application, so there is no issuer
or audience to tell apart, and a claim that verifiers are told to ignore only
invites careless verification. If your application needs more than the user
ID, look it up by `sub`.

The `kid` header names the signing key, and matches a `kid` in the JWKS.

## What to verify

A correct verifier checks all of these:

1. **The algorithm is `EdDSA`.** Pin it. Trusting the algorithm the token
   names is the classic JWT vulnerability: a token with `"alg": "none"`, or
   with an HMAC algorithm that uses the public key as its secret, must be
   rejected.
2. **The signature is valid** under the public key whose `kid` matches the
   token's header.
3. **`exp` is present and in the future.** A token without `exp` must be
   rejected, not treated as never expiring.
4. **`sub` is a UUID.**

Mainstream JWT libraries do 2 and 3 for you; 1 usually needs an explicit
option, and 4 is up to you.

`go-auth` verifies its own tokens the same way, split over two files, and both
can serve as a reference: `internal/infra/jwt/ed25519/signer.go` pins the
algorithm, picks the key by `kid` and requires `exp`, which is checks 1 to 3,
and `internal/infra/jwt/access_token_service.go` parses `sub` into a user ID,
which is check 4. That code is not reachable over HTTP: the service has no
endpoint that verifies an access token, because verifying it in your own
service is the point.

## Fetching and caching keys

```
GET /.well-known/jwks.json
```

```json
{
  "keys": [
    {
      "kty": "OKP", "crv": "Ed25519", "use": "sig", "alg": "EdDSA",
      "x": "11qYAYKxCrfVS_7TyWQHOg7hcvPapiMlrwIaaPcHURo",
      "kid": "NzbLsXh8uDCcd-6MNwXF4W_7noWXFZAfHkxZsRGC9Xs"
    }
  ]
}
```

`x` is the raw 32-byte Ed25519 public key, in base64url without padding.

Cache the document: fetching it on every request defeats the purpose. The
response has no cache headers, so choose the cache lifetime yourself. Cache
it for a few minutes, fetch it again when a token has a `kid` you do not
know, and limit how often that refetch can happen, so tokens with random
`kid` values cannot force a flood of requests.

A key is in the document a day before it starts signing, and stays for 25
hours after it stops, longer than any access token lives. So a verifier that
fetches the document at least every 23 hours never meets a valid token whose
`kid` it lacks, and no valid token loses its key. Restarts change nothing. See
[Signing keys](../architecture/tokens.md#signing-keys).

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

## Revocation is not immediate

This is the trade-off of verifying tokens yourself.

An access token is checked without asking `go-auth`, so nothing can invalidate
one before its `exp`: not reuse detection, not a revoked session. When a
session is revoked, its refresh token stops working at once, but access tokens
already issued stay valid for the rest of their lifetime.

`ACCESS_TOKEN_TTL` is therefore how long revocation takes. With the default of
30 minutes, a stolen access token keeps working for up to 30 minutes after
its session is revoked. Shorten it if that is too long for your application;
the cost is more refresh calls.

If your application needs revocation to take effect at once, verifying tokens
yourself is the wrong tool: every request would need a check against the
server, and `go-auth` has no endpoint for it.
