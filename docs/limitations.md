# Limitations

What `go-auth` does not do, and the trade-offs to know before deploying it.
Everything here describes the current behaviour: none of it is a defect
waiting for a fix.

## Out of scope

`go-auth` checks a username and password and issues tokens. It does not
provide:

- email, OAuth or passwordless login;
- password reset or password change;
- logout: a session ends only when its refresh token expires, or when
  [reuse detection](architecture/usecases/refresh.md#reuse-detection) revokes it;
- account or session management, such as listing sessions or suspending an
  account;
- authorisation: there are no roles or scopes, and an access token carries
  only the user ID;
- multi-tenancy: one deployment serves one application;
- an audit trail: authentication events are log lines, not stored records;
- health or readiness endpoints, or metrics;
- TLS, CORS or rate limiting (see
  [Run it behind a reverse proxy](#run-it-behind-a-reverse-proxy)).

## Deployment

### One replica, and a restart invalidates access tokens

The key that signs access tokens is generated at startup and held only in
memory. As a result:

- **every restart invalidates every outstanding access token.** Your services
  reject them until each client refreshes. Refresh tokens are stored in the
  database and survive the restart, so refreshing works.
- **`go-auth` runs as a single replica.** Two instances sign with different
  keys, and each publishes only its own, so a token issued by one fails
  verification against the other's JWKS.

### Key rotation has no overlap

The signing key is replaced every 7 days, and on every restart. Only the
current key is published:

- the previous key leaves `/.well-known/jwks.json` at once, so tokens it
  signed stop verifying before they expire;
- the new key is published only when it starts signing, so a verifier with a
  cached JWKS rejects fresh tokens until it fetches the document again.

Verifiers should refetch the JWKS on an unknown `kid`
([how](api/token-verification.md#fetching-and-caching-keys)), and clients
should refresh when a service answers `401`.

### Run it behind a reverse proxy

`go-auth` serves plain HTTP, with no rate limiting and no CORS handling. Put
it behind a reverse proxy that:

- terminates TLS;
- rate limits the three `POST` endpoints. Nothing else limits password
  guessing, apart from the cost of a bcrypt comparison;
- exposes only the four endpoints.

A proxy limits by client address, not by account, so guesses against one
account spread over many addresses are not limited.

A browser sends a cross-site `POST` as `text/plain`, a form or a URL-encoded
body without asking first, but asks (a CORS preflight) before sending
`application/json`, and `go-auth` never allows it. Answering `415` to anything
but `application/json` is therefore what stops a page on another site from
sending logins and registrations through its visitors' browsers, one client
address each, past the proxy's rate limit. A proxy that answers those
preflights for these endpoints, for origins you do not control, undoes that
([why](development/decisions/0051-post-endpoints-require-application-json.md)).

### Refresh token rows are never deleted

Every login and every refresh inserts a row into `refresh_tokens`, and nothing
deletes them, so the table grows as long as the deployment runs.

A spent token is what reuse detection looks for. If one is deleted, replaying
it gets "not found" instead of revoking the session. Only the tokens of a
session that is revoked, or whose latest token has expired, are safe to
delete. Deleting the session deletes its tokens with it.

## Sessions and tokens

### Access tokens cannot be revoked

Revoking a session stops its refresh token at once, but access tokens already
issued stay valid until they expire. `ACCESS_TOKEN_TTL` is how long a revoked
session keeps working. See
[Revocation is not immediate](api/token-verification.md#revocation-is-not-immediate).

A refresh that is already running when its session is revoked can also still
finish and return new tokens. It checks the session once, at the start. The
result is the same as if the refresh had finished just before the revocation,
which is a valid order for two concurrent requests. The refresh token it
returns is useless, because its session is revoked.

### Sessions have no maximum age

Each refresh issues a token with a fresh `REFRESH_TOKEN_TTL`, so a client that
refreshes regularly stays signed in indefinitely. `REFRESH_TOKEN_TTL` limits
how long a session can go unused, not how long it can last.

### A repeated refresh logs the user out

A refresh token can be used once. When it arrives a second time, the service
cannot tell the legitimate client from someone who copied the token, so it
revokes the session
([why](architecture/usecases/refresh.md#reuse-detection)). A legitimate client
also sends a token twice sometimes:

- **the response is lost.** The refresh succeeds, but the response never
  arrives, because of a timeout or a dropped connection. The client still
  holds the spent token, and retrying with it revokes the session.
- **several tabs refresh.** Tabs or app instances that share storage refresh
  the same token at the same time. One succeeds, and the others revoke the
  session, which signs every tab out.

Either way the user has to sign in again. Clients avoid it by following the
[client obligations](architecture/usecases/refresh.md#client-obligations).

## Login timing after a `BCRYPT_COST` change

When the account does not exist, login compares the password against a dummy
hash, so a missing account takes as long as a wrong password. The dummy hash
is created at startup with the current `BCRYPT_COST`. A stored hash keeps the
cost it was created with, and nothing rehashes it.

Once `BCRYPT_COST` changes, accounts created before the change no longer take
as long as a missing account. Raising the cost from 12 to 14 makes a missing
account about four times slower than an old one, and lowering it makes the
old accounts slower. Either way, response time shows which of those usernames
exist. Accounts created after the change are not affected. Raising the cost
also leaves the old hashes as easy to crack offline as before.

This matters little in practice: registration already answers
`409 USERNAME_ALREADY_EXISTS` for a taken username, so whether an account
exists is not secret. If it matters for your deployment, choose `BCRYPT_COST`
before the first user registers, and keep it.
