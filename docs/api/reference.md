# API reference

The endpoints, with JSON bodies. The same contract, machine-readable, is
[openapi.yaml](openapi.yaml) (OpenAPI 3.1); a change to the contract updates
both.

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/v1/auth/register` | Create a user |
| `POST` | `/v1/auth/login` | Exchange credentials for a token pair |
| `POST` | `/v1/auth/change-password` | Replace a password, given the current one, and revoke every session |
| `POST` | `/v1/auth/refresh` | Exchange a refresh token for a new pair |
| `POST` | `/v1/auth/verify` | Check an access token, for a service that does not verify it itself |
| `POST` | `/v1/auth/me` | Return the user an access token was issued to |
| `GET` | `/.well-known/jwks.json` | Public keys for verifying access tokens |

## Conventions

**Versioning.** The auth endpoints start with `/v1`. The JWKS has no version,
because [RFC 8615](https://www.rfc-editor.org/rfc/rfc8615) fixes its path.

**Content type.** The `POST` endpoints require `Content-Type: application/json`,
matched regardless of case, with any parameters (`; charset=utf-8`) ignored.
Anything else, a missing header included, is answered
`415 unsupported_media_type` before the body is read. `curl -d` sends a form
type, so add `-H 'Content-Type: application/json'` or use `--json`. Every
response body is JSON, except the redirects below, and the empty body of a
`500` whose response could not be encoded.

**Methods and routes.** The auth endpoints accept `POST`, and the JWKS `GET`
and `HEAD`. Another method on a known path is `405 method_not_allowed`, with an
`Allow` header naming the accepted ones; an unknown path is
`404 route_not_found`. A path not in canonical form (`//v1/auth/login`, or one
with `..` segments) gets a `307` redirect to its cleaned form from Go's
`http.ServeMux`: a short HTML body for a `GET`, an empty one otherwise.

**Rate limiting.** When the operator turns it on, the endpoints listed under
[`RATE_LIMIT`](../configuration.md#rate_limit) allow each client address a
number of requests, then answer `429 too_many_requests` with a `Retry-After`
header in seconds. Nothing ran ([why](../system.md#rate-limiting)): send the
same request again after the delay, and on `/v1/auth/refresh` the same token,
which was not spent.

**Timestamps.** RFC 3339, in UTC, to the whole second
(`2026-09-08T12:00:00Z`). A fraction of a second is dropped, never rounded up,
so an `expires_at` is never later than the real expiry.

**Errors.** One shape, described in the [Error model](errors.md). Besides the
errors listed for each endpoint, every `POST` endpoint can answer
`400 invalid_json_body`, `413 request_body_too_large`,
`415 unsupported_media_type` and `500 internal_server_error`, and a rate-limited
one `429 too_many_requests`.

### Request bodies

The body of a `POST` is one JSON object. An empty body, malformed or truncated
JSON, a value that is not an object (`[]`, `"text"`), a field of the wrong type
(`{"username": 1}`), and invalid UTF-8 anywhere in the object, in a field that
is ignored too, are all `400 invalid_json_body`. Invalid UTF-8 is rejected
rather than replaced because a replaced password would be stored that way:
every invalid byte becomes the same `U+FFFD`, so a body sent in Latin-1 would
make `senhaçã` and `senhaõé` one password.

A body is decoded up to its first 64 KiB (65,536 bytes): one whose first JSON
value does not end within them is `413 request_body_too_large`. The limit is not
a setting.

**Unknown fields are ignored**, which is what lets a field be added without
breaking clients. A misspelled field is therefore not reported as such: the
field it was meant to be arrives empty, and is reported instead
(`{"usernme": "alice"}` at registration is `too_short` on `username`). A missing
field is read as `""`.

The decoder also tolerates the following. Send one well-formed object, in valid
UTF-8, with each field once and in lower case, and do not depend on these:

- a body of `null` reads as `{}`;
- a field set to `null` is read as `""`, like a missing one;
- anything after the first JSON value is ignored;
- field names match regardless of case, and a field sent twice keeps the last
  value;
- an escaped lone surrogate such as `\ud800` becomes `U+FFFD`, which in a
  password has consequences
  ([Limitations](../system.md#a-lone-surrogate-in-a-password-becomes-ufffd)).

---

## POST /v1/auth/register

Creates a user. It does **not** log the user in: call `/v1/auth/login` next.

```json
{ "username": "alice", "password": "correct-horse" }
```

| Field | Rules |
|---|---|
| `username` | 3–32 code points; only `a-z`, `0-9`, `.`, `_`, `-`, after lower-casing |
| `password` | at least 8 code points, at most 72 bytes; any character |

Usernames are lower-cased before they are checked and stored, so `Alice` and
`alice` are the same user. Spaces are not trimmed: a leading or trailing space
is a character outside the allowed set. Passwords are stored and compared
exactly as sent, with no trimming and no Unicode normalisation. The reasons:
[Usernames and passwords](../system.md#usernames-and-passwords).

### Response: `201 Created`

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

### Errors

| Status | Code | Cause |
|---|---|---|
| `409` | `username_already_exists` | The username is taken |
| `422` | `validation_failed` | The username or the password breaks a rule |

A `422` names every field at fault, so one request reports every problem:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "The input failed validation.",
    "fields": [
      { "field": "username", "code": "too_short", "details": { "min": 3, "unit": "code_point" } },
      { "field": "password", "code": "too_short", "details": { "min": 8, "unit": "code_point" } }
    ]
  }
}
```

---

## POST /v1/auth/login

Checks the credentials and opens a new session.

```json
{ "username": "alice", "password": "correct-horse" }
```

### Response: `200 OK`

```json
{
  "access_token": {
    "value": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9...",
    "expires_at": "2026-09-08T12:30:00Z",
    "expires_in": 1799
  },
  "refresh_token": {
    "value": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01",
    "expires_at": "2026-09-15T12:00:00Z",
    "expires_in": 604800
  }
}
```

`expires_in` is the whole seconds from the moment the service issued the token
to its expiry, rounded down. If you cannot trust your own clock, count it from
when the response arrives, and refresh a little early: the time on the network
is not deducted.

`access_token.value` is a JWT: send it to your services, and verify it there
([how](token-verification.md)). `refresh_token.value` is 32 random bytes in
base64url, meaningless outside this service. Treat it as a password: keep it
private, never log it, never put it in a URL.

### Errors

| Status | Code | Cause |
|---|---|---|
| `401` | `invalid_credentials` | Any failure: an unknown username, a wrong password, or a value that breaks the registration rules |

Login never answers `422`
([why](../system.md#deliberately-vague-answers)).

---

## POST /v1/auth/change-password

Replaces a user's password, given the current one, and signs the user out of
every session. It returns no tokens: log in again with the new password.

```json
{ "username": "alice", "current_password": "correct-horse", "new_password": "battery-staple" }
```

| Field | Rules |
|---|---|
| `username`, `current_password` | as at login |
| `new_password` | as a registration's password, and not the same as `current_password` |

`new_password` is compared with `current_password` byte for byte, and only with
it: a password used before is accepted again.

### Response: `200 OK`

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

The old password stops working at once, and **every session of the user is
revoked**: every refresh token issued before the change, on any device, is
answered `401 invalid_token` from then on. Access tokens already issued stay
valid until they expire
([Limitations](../system.md#access-tokens-cannot-be-revoked)).

### Errors

| Status | Code | Cause |
|---|---|---|
| `401` | `invalid_credentials` | The username or the current password was rejected, for any reason, as at login |
| `422` | `validation_failed` | `new_password` breaks a rule, or is the same as `current_password` (`same_as_current`) |

A `422` names only `new_password`, and comes before the account is looked up:

```json
{
  "error": {
    "code": "validation_failed",
    "message": "The input failed validation.",
    "fields": [
      { "field": "new_password", "code": "same_as_current", "details": {} }
    ]
  }
}
```

What happens to the logins and changes running at the same moment:
[Changing the password](../system.md#changing-the-password).

---

## POST /v1/auth/refresh

Exchanges a refresh token for a new access token and a new refresh token.

```json
{ "refresh_token": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01" }
```

### Response: `200 OK`

The same shape as the login response. **The refresh token you sent is now
spent**: replace your stored copy with the new one.

### Errors

| Status | Code | Cause |
|---|---|---|
| `401` | `invalid_token` | Any failure: unknown, expired, already used, or its session revoked |

**Reuse detection.** A token that was already used revokes its whole session
before the `401`: the session's current refresh token stops working too, and
the user has to log in again. Two concurrent refreshes with the same token
count as reuse ([why](../system.md#reuse-detection)).

### Client obligations

1. Store the refresh token privately and durably.
2. After a successful refresh, replace it with the new one at once.
3. Never retry a refresh with the old token: if the response was lost, the
   retry trips reuse detection and logs the user out. Log in again instead.
4. Never run two refreshes with the same token at once: across tabs, threads
   or instances sharing storage, only one may be in flight.
5. On a `401`, discard the token and send the user to log in.

A `429` is the exception to point 3: the token was not spent, so wait for
`Retry-After` and send the same one.

---

## POST /v1/auth/verify

Checks an access token and returns the user it was issued to, for a service
that would rather not verify access tokens itself. It makes the checks a
correct local verifier makes, with the keys the JWKS publishes, and nothing
else ([Letting go-auth verify](token-verification.md#letting-go-auth-verify)).

```json
{ "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..." }
```

### Response: `200 OK`

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

The answer holds until `expires_at`: nothing makes an access token invalid
sooner, not even revoking its session.

### Errors

| Status | Code | Cause |
|---|---|---|
| `401` | `invalid_token` | Any failure of the [checks](../system.md#checking-an-access-token), an expired token included |

---

## POST /v1/auth/me

Checks an access token and returns the user it was issued to, for a client
that holds the token and needs more than the user ID it carries, such as the
username to show. It checks the token as [verify](#post-v1authverify) does,
then reads the user.

```json
{ "access_token": "eyJhbGciOiJFZERTQSIsImtpZCI6IjRxNi4uLiIsInR5cCI6IkpXVCJ9..." }
```

### Response: `200 OK`

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

The same `user` as register returns. As with verify, revoking the token's
session does not change the answer: the user is returned until the token
expires.

### Errors

| Status | Code | Cause |
|---|---|---|
| `401` | `invalid_token` | Any failure: whatever verify rejects, and a token whose user no longer exists |

---

## GET /.well-known/jwks.json

The public keys that verify access tokens, as a
[JWK Set](https://www.rfc-editor.org/rfc/rfc7517). It holds no secrets and needs
no authentication.

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

It usually holds one or two keys, and has no cache headers. How to cache it:
[Fetching and caching keys](token-verification.md#fetching-and-caching-keys).
