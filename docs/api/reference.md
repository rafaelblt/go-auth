# API reference

Four endpoints. All request and response bodies are JSON.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/auth/register` | Create a user |
| `POST` | `/v1/auth/login` | Exchange credentials for a token pair |
| `POST` | `/v1/auth/refresh` | Exchange a refresh token for a new pair |
| `GET` | `/.well-known/jwks.json` | Public keys for verifying access tokens |

## Conventions

**Versioning.** The auth endpoints start with `/v1`. The JWKS endpoint has no
version, because [RFC 8615](https://www.rfc-editor.org/rfc/rfc8615) fixes its
path.

**Content type.** Every response the service writes is
`Content-Type: application/json`. The three `POST` endpoints require
`Content-Type: application/json`, matched regardless of case, with any
parameters (`; charset=utf-8`) allowed and ignored. Anything else, including no
header at all, is answered `415 UNSUPPORTED_MEDIA_TYPE` before the body is
read. `curl -d` sends `application/x-www-form-urlencoded`, so add
`-H 'Content-Type: application/json'` or use `--json`.

**Methods.** Each route accepts one method. A `GET` to `/v1/auth/login`
returns `405 Method Not Allowed`, not `404`.

**Unknown routes return plain text.** A `404` for an unknown path, and a `405`
for a known path with the wrong method, come from Go's `http.ServeMux`, with a
`text/plain` body instead of the JSON error shape. Do not assume every error
response has a JSON body.

**Timestamps.** RFC 3339, in UTC: `2026-09-08T12:00:00Z`.

**Errors.** One shape, an `error` object with a `code`, described in
[Error model](errors.md).

### Request bodies

The body of a `POST` is one JSON object. An empty body, malformed or truncated
JSON, a value that is not an object (`[]`, `"text"`), and a field of the wrong
type (`{"username": 1}`) are all `400 INVALID_JSON_BODY`.

**Unknown fields are ignored.** That is a rule, and it is what lets a field be
added without breaking clients. It also means a misspelled field name is not
reported as such: the field it was meant to be arrives empty, and is reported
instead. `{"usernme": "alice"}` at registration is `TOO_SHORT` on `username`.

A missing field and a field set to `null` are both read as `""`.

The decoder also tolerates the following. Send one well-formed object in valid
UTF-8, with each field once and in lower case, and do not depend on these:

- A body of `null` reads as `{}`: registration answers `422`, login and refresh
  `401`.
- Anything after the first JSON value is ignored.
- Field names match regardless of case: `USERNAME` fills `username`.
- A field sent twice keeps the last value, also when the two differ in case.
- Invalid UTF-8 in a string, and an escaped lone surrogate such as `\ud800`,
  become `U+FFFD` without an error. In a password, each one replaced counts as
  3 bytes toward the 72-byte limit.

---

## POST /v1/auth/register

Creates a user with a username and a password. It does **not** log the user
in and returns no tokens: call `/v1/auth/login` next.

### Request

```json
{
  "username": "alice",
  "password": "correct-horse"
}
```

| Field | Type | Rules |
|---|---|---|
| `username` | string | 3–32 code points; only `a-z`, `0-9`, `.`, `_`, `-` |
| `password` | string | at least 8 code points, at most 72 bytes |

Usernames are lower-cased before they are checked and stored, so `Alice` and
`alice` are the same user, stored as `alice`. Spaces are not trimmed: a
leading or trailing space is a character outside the allowed set.

Passwords are stored and compared exactly as sent: no trimming, no Unicode
normalisation, and any character is allowed. The one change happens while the
body is decoded: invalid UTF-8 becomes `U+FFFD` (see
[Request bodies](#request-bodies)). The reasons for these rules are in the
[Domain model](../architecture/domain/user.md#username).

### Response: `200 OK`

```json
{
  "user": {
    "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37",
    "username": "alice",
    "status": "active",
    "created_at": "2026-09-08T12:00:00Z",
    "updated_at": "2026-09-08T12:00:00Z"
  }
}
```

`status` is always `active`.

### Errors

| Status | Code | Cause |
|---|---|---|
| `400` | `INVALID_JSON_BODY` | The body is not a JSON object with fields of the expected types |
| `409` | `USERNAME_ALREADY_EXISTS` | The username is taken |
| `415` | `UNSUPPORTED_MEDIA_TYPE` | The `Content-Type` is not `application/json` |
| `422` | `VALIDATION_FAILED` | The username or the password is invalid |

A validation error names every field at fault, so one request reports every
problem:

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "The input failed validation.",
    "fields": [
      { "field": "username", "code": "TOO_SHORT", "details": { "min": 3, "unit": "code_point" } },
      { "field": "password", "code": "TOO_SHORT", "details": { "min": 8, "unit": "code_point" } }
    ]
  }
}
```

Usernames are unique even under concurrency: of any number of registrations
of one username sent at the same moment, exactly one creates the account, and
the others get `409 USERNAME_ALREADY_EXISTS`, as they would one after the
other.

---

## POST /v1/auth/login

Checks the credentials and opens a new session, returning an access token and
a refresh token.

### Request

```json
{
  "username": "alice",
  "password": "correct-horse"
}
```

### Response: `200 OK`

```json
{
  "access_token": {
    "value": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-09-08T12:30:00Z"
  },
  "refresh_token": {
    "value": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01",
    "expires_at": "2026-09-15T12:00:00Z"
  }
}
```

`access_token.value` is a JWT. Send it to your own services, and verify it
there ([how](token-verification.md)).

`refresh_token.value` is 32 random bytes in base64url, not a JWT. It means
nothing outside this service, which stores only its SHA-256 hash. Treat it as
a password: keep it private, never log it, never put it in a URL.

### Errors

| Status | Code | Cause |
|---|---|---|
| `400` | `INVALID_JSON_BODY` | The body is not a JSON object with fields of the expected types |
| `401` | `INVALID_CREDENTIALS` | Any other failure |
| `415` | `UNSUPPORTED_MEDIA_TYPE` | The `Content-Type` is not `application/json` |

Every login failure gets the same response: a malformed username, an unknown
username, a user without a password, and a wrong password. Login never returns
`422`. Nothing in the response, or in how long it takes, tells an unknown
username from a wrong password; see
[Deliberately vague errors](errors.md#deliberately-vague-errors).

---

## POST /v1/auth/refresh

Exchanges a valid refresh token for a new access token and a new refresh
token. The token sent is spent in the process.

### Request

```json
{
  "refresh_token": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01"
}
```

### Response: `200 OK`

The same shape as the login response. **The refresh token you sent is now
spent.** Replace your stored copy with the new one: sending the old one again
revokes the whole session.

### Errors

| Status | Code | Cause |
|---|---|---|
| `400` | `INVALID_JSON_BODY` | The body is not a JSON object with fields of the expected types |
| `401` | `INVALID_TOKEN` | Any other failure |
| `415` | `UNSUPPORTED_MEDIA_TYPE` | The `Content-Type` is not `application/json` |

Every failure gets the same response: a token that never existed, one that
expired, one already used, and one whose session was revoked. A caller cannot
tell which it was, or whether reuse detection fired.

**Reuse detection.** If the token was already used, the service revokes its
whole session before answering `401`. The session's current refresh token
stops working too, and the user has to log in again. Two concurrent refreshes
with the same token count as reuse: one succeeds, the other revokes the
session. Clients must follow the
[client obligations](../architecture/usecases/refresh.md#client-obligations) to
avoid logging their own users out.

---

## GET /.well-known/jwks.json

Publishes the public keys that verify access tokens, as a
[JWK Set](https://www.rfc-editor.org/rfc/rfc7517). It holds no secrets and
needs no authentication.

### Response: `200 OK`

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
| `kty` | Key type: always `OKP` (octet key pair) |
| `crv` | Curve: always `Ed25519` |
| `x` | The public key, base64url without padding |
| `use` | Always `sig`: the key verifies signatures |
| `alg` | Always `EdDSA` |
| `kid` | Key ID: the RFC 7638 thumbprint of the key |

A token's `kid` header matches the `kid` of the key that signed it, so a
verifier can pick the right key without trying each one.

The document holds exactly one key, the one signing now, and the response has
no cache headers. How to cache it, and what happens when the key rotates:
[Fetching and caching keys](token-verification.md#fetching-and-caching-keys).
