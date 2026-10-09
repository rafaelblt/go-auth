# Limitations

What `go-auth` does not do, and the trade-offs to know before deploying it.
Everything here describes the current behaviour: none of it is a defect
waiting for a fix.

## Out of scope

`go-auth` checks a username and password and issues tokens. It does not
provide:

- email, OAuth or passwordless login;
- password reset;
- logout: a session ends only when its refresh token expires, or when
  [reuse detection](architecture/usecases/refresh.md#reuse-detection) or a
  [password change](architecture/usecases/change-password.md) revokes it;
- account or session management, such as listing sessions or suspending an
  account;
- authorisation: there are no roles or scopes, and an access token carries
  only the user ID;
- multi-tenancy: one deployment serves one application;
- an audit trail: authentication events are log lines, not stored records;
- health or readiness endpoints, or metrics;
- TLS or CORS (see
  [Run it behind a reverse proxy](#run-it-behind-a-reverse-proxy)).

## Deployment

### Signing keys are stored in the database

The keys that sign access tokens are rows of the `signing_keys` table, so
every instance signs and publishes the same keys, and a restart changes
nothing ([how](architecture/tokens.md#signing-keys)). As a result:

- **they are stored unencrypted unless
  [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key)
  is set.** Without it, whoever reads the database, or a backup of it, can
  sign access tokens that every verifier accepts while that key is published.
  A key is published for about 9 days from when it is added: a day before it
  signs, 7 days signing, and 25 hours after. So, while rotation runs, a backup
  holds no usable key after that. Protect the database and its backups like
  the keys they hold. With it set, the database holds only sealed keys, apart
  from those stored before it was set, until they are deleted. v2 requires it.
- **the encryption key cannot be rotated.** Changing or removing it leaves the
  stored keys unreadable, and the service does not start. The way out deletes
  them, which invalidates outstanding access tokens once
  ([how](configuration.md#signing_key_encryption_key)).
- **instances go by their own clocks** for when a key starts signing and when
  an old one is dropped. The margins absorb up to an hour of skew, so keep the
  clocks synchronised.

### Run it behind a reverse proxy

`go-auth` serves plain HTTP, with no CORS handling, and its rate limiting is
off unless [`RATE_LIMIT`](configuration.md#rate_limit) turns it on. Put it
behind a reverse proxy that:

- terminates TLS;
- rate limits register, login, change password and refresh when
  `RATE_LIMIT` is `off`. Nothing else limits password guessing then, apart
  from the cost of a bcrypt comparison;
- exposes only the six endpoints.

With the service's own rate limiting on, set
[`TRUSTED_PROXIES`](configuration.md#trusted_proxies) to the proxy, or every
client is counted as the proxy and they all share its allowance. The same
holds for an application server that calls the service for its users: it
has to forward their addresses, and be listed too.

A rate limit by client address, the proxy's or the service's, is not a limit
by account, so guesses against one account spread over many addresses are not
limited.

A browser sends a cross-site `POST` as `text/plain`, a form or a URL-encoded
body without asking first, but asks (a CORS preflight) before sending
`application/json`, and `go-auth` never allows it. Answering `415` to anything
but `application/json` is therefore what stops a page on another site from
sending logins and registrations through its visitors' browsers, one client
address each, past any per-address rate limit. A proxy that answers those
preflights for these endpoints, for origins you do not control, undoes that
([why](development/decisions/0051-post-endpoints-require-application-json.md)).

### Rate limiting is per address and per process

When [`RATE_LIMIT`](configuration.md#rate_limit) turns it on, the service
counts requests per client address:

- **per address, not per account.** Guesses against one account spread over
  many addresses are not limited.
- **an address can be many users.** Clients behind one NAT share an
  allowance, and an IPv6 /64 counts as one address. `relaxed` exists for
  that. The users of an application server that calls the service for them
  share its allowance too, unless it forwards their addresses
  ([`TRUSTED_PROXIES`](configuration.md#trusted_proxies)). v2 cannot turn
  rate limiting off, so by then such a server has to forward them.
- **in memory.** The counts live in the process, so a restart restores every
  allowance, and each instance counts on its own: N instances allow N times
  the limit.
- **only `X-Forwarded-For`** is read, and only from a
  [trusted proxy](configuration.md#trusted_proxies); `Forwarded` and
  `X-Real-IP` are ignored.
- **off by default in v1.** `off` is deprecated, and v2 removes it.

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
session keeps working, after reuse detection and after a password change
alike. [`POST /v1/auth/verify`](api/reference.md#post-v1authverify)
does not change that: it makes the checks a local verifier makes. See
[Revocation is not immediate](api/token-verification.md#revocation-is-not-immediate).

A refresh that is already running when its session is revoked can also still
finish and return new tokens. It checks the session once, at the start. The
result is the same as if the refresh had finished just before the revocation,
which is a valid order for two concurrent requests. The refresh token it
returns is useless, because its session is revoked.

### A login running during a password change outlives it

[Changing the password](api/reference.md#post-v1authchange-password) revokes
the sessions of the user that exist when the change commits. A login that read
the old password before that, and opens its session after, is not reached: its
session works, although it was opened with a password that no longer does.
The window is about as long as a bcrypt comparison, and only someone who
already holds the old password can use it, at the moment the user changes it.
Closing it would take a guarded write in login: saving the session only while
the stored password is still the one it checked.

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

When the account does not exist, login and change password compare the
password against a dummy hash, so a missing account takes as long as a wrong
password. The dummy hash is created at startup with the current
`BCRYPT_COST`. A stored hash keeps the cost it was created with until the user
changes the password: nothing rehashes it.

Once `BCRYPT_COST` changes, accounts created before the change no longer take
as long as a missing account. Raising the cost from 12 to 14 makes a missing
account about four times slower than an old one, and lowering it makes the
old accounts slower. Either way, response time shows which of those usernames
exist. Accounts created after the change, or whose password was changed after
it, are not affected. Raising the cost also leaves the old hashes as easy to
crack offline as before.

This matters little in practice: registration already answers
`409 username_already_exists` for a taken username, so whether an account
exists is not secret. If it matters for your deployment, choose `BCRYPT_COST`
before the first user registers, and keep it.

## A lone surrogate in a password becomes `U+FFFD`

A request body in invalid UTF-8 is rejected, but an escaped lone surrogate
such as `\ud800` is valid UTF-8 on the wire, and the JSON decoder replaces it
with `U+FFFD` without an error
([Request bodies](api/reference.md#request-bodies)). In a password, that
replacement is what gets hashed and stored, so:

- **different passwords can be the same one.** Every lone surrogate becomes
  the same character: `ab\ud800` and `ab\udfff` log in to the same account.
- **the password is longer than sent.** Each replacement counts as 3 bytes
  toward the 72-byte limit.
- **the account depends on the replacement.** A password stored this way
  logs in only while lone surrogates keep becoming `U+FFFD`. A decoder that
  rejected them would lock that user out.

A client sends one only when its string is not well-formed Unicode, such as a
JavaScript string cut through the middle of an emoji: `JSON.stringify` then
escapes the half it kept. Clients avoid it by sending well-formed strings.
