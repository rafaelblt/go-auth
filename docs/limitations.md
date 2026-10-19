# Limitations

What `go-auth` does not do, and the trade-offs to know before deploying it.
Everything here is current behaviour, not a defect waiting for a fix.

## Out of scope

`go-auth` checks a username and password and issues tokens. It does not
provide:

- email, OAuth or passwordless login, or password reset;
- logout: a session ends only when its refresh token expires, or when
  [reuse detection](architecture/usecases.md#reuse-detection) or a password
  change revokes it;
- account or session management, such as listing sessions or suspending an
  account;
- authorisation: there are no roles or scopes, and an access token carries
  only the user ID;
- a password policy beyond its length: no composition rules, and no check
  against lists of common or breached passwords;
- multi-tenancy: one deployment serves one application;
- an audit trail: authentication events are log lines, not stored records;
- health or readiness endpoints, or metrics;
- TLS or CORS.

## Deployment

### Run it behind a reverse proxy

`go-auth` serves plain HTTP, and its rate limiting is off unless
[`RATE_LIMIT`](configuration.md#rate_limit) turns it on. Put it behind a
reverse proxy that:

- terminates TLS;
- rate limits register, login, change password, refresh and me while
  `RATE_LIMIT` is `off`: nothing else limits password guessing then, beyond
  the cost of a bcrypt comparison;
- exposes only the seven endpoints.

With the service's own rate limiting on, set
[`TRUSTED_PROXIES`](configuration.md#trusted_proxies) to the proxy, or every
client is counted as the proxy. A proxy that answers CORS preflights for these
endpoints, for origins you do not control, undoes what the `415` protects
([why](architecture/http.md#content-type)).

### Rate limiting is per address and per process

- **Per address, not per account.** Guesses against one account spread over
  many addresses are not limited, whether the proxy or the service limits them.
- **An address can be many users.** Clients behind one NAT share an
  allowance, and so does an IPv6 /64; `relaxed` exists for that. The users of
  an application server that calls the service for them share its allowance
  too, unless it forwards their addresses.
- **In memory.** A restart restores every allowance, and each instance counts
  on its own: N instances allow N times the limit.
- **Off by default.** `off` is deprecated, and v2 removes it.

### Signing keys are stored in the database

The keys that sign access tokens are rows of `signing_keys`, so every instance
uses the same ones ([Signing keys](architecture/tokens.md#signing-keys)). As a
result:

- **they are unencrypted unless
  [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key)
  is set.** Without it, whoever reads the database, or a backup of it, can sign
  access tokens that every verifier accepts while that key is published, about
  9 days from when it was added. Protect the database and its backups like the
  keys they hold.
- **the encryption key cannot be changed in place.** Changing it means
  [replacing every signing key](configuration.md#replacing-the-signing-keys),
  which invalidates the access tokens already issued.
- **instances go by their own clocks** to decide which key signs and when an
  old one is dropped. The windows absorb up to an hour of skew, so keep the
  clocks synchronised.

### A session in use keeps every refresh token

Every login inserts a session and a refresh token, and every refresh a refresh
token. The [session cleanup](architecture/usecases.md#session-cleanup) deletes a
session, with all its tokens, once it has been dead for
[`SESSION_RETENTION`](configuration.md#session_retention), and never a token on
its own ([why](architecture/persistence.md#deleting-dead-sessions)). A session
that stays in use therefore keeps every token it was issued, and, with
[no maximum age](#sessions-have-no-maximum-age), keeps adding them: a client
that refreshes every 30 minutes adds 48 rows a day for as long as it stays
signed in.

With `SESSION_RETENTION` set to `off`, nothing is deleted, and both tables grow
for as long as the deployment runs.

### The session cleanup can fall behind

The [session cleanup](architecture/usecases.md#session-cleanup) runs on each
instance as a background task, whose first run comes one interval after the
instance starts, and each run deletes a bounded number of sessions
([Background tasks](architecture/startup.md#background-tasks)). As a result:

- **an instance that lives less than an interval never runs it.** When every
  instance is replaced sooner, by frequent deploys, autoscaling or a platform
  that stops idle instances, nothing is deleted, and both tables grow as they
  do with `SESSION_RETENTION` set to `off`.
- **a deployment that opens sessions faster than its instances delete them
  falls behind**, and its dead sessions pile up until the rate drops. Nothing
  in the logs shows it: the line a run writes says how long it took, not how
  many sessions it deleted ([Background tasks](logging.md#background-tasks)).

## Accounts

### Registration reveals whether a username exists

Registration answers `409 username_already_exists` for a taken username, so
anyone can learn whether an account exists, and, unless rate limiting is on, as
often as they like. Login and change password do not tell an unknown username
from a wrong password ([Deliberately vague errors](api/errors.md#deliberately-vague-errors)),
so what they keep secret is whether a password is right, not whether its
account exists.

### Whoever knows the password can lock the owner out

[Changing the password](api/reference.md#post-v1authchange-password) asks for
the username and the current password, and nothing else: no second factor, no
confirmation sent elsewhere, no notice to the owner. The change revokes every
session, the owner's included, and the service offers no way back: no password
reset, no account management. Getting the account back is up to whatever the
deployment builds around the service. It works the other way too: an owner who
changes the password signs out whoever else held it.

### Login timing after a `BCRYPT_COST` change

A missing account costs a bcrypt comparison against a dummy hash made at
startup, with the current `BCRYPT_COST`, so it takes as long as a wrong
password. A stored hash keeps the cost it was made with until its user changes
the password. Once `BCRYPT_COST` changes, the accounts hashed before no longer
take as long as a missing one, and response time shows which of those
usernames exist. Raising the cost also leaves those older hashes as easy to
crack offline as before. To avoid both, choose `BCRYPT_COST` before the first
user registers, and keep it.

## Sessions and tokens

### Access tokens cannot be revoked

Revoking a session, by reuse detection or by a password change, stops its
refresh token at once, but the access tokens already issued stay valid until
they expire: they hold everything needed to check them, not even their session.
`ACCESS_TOKEN_TTL` is therefore how long a revoked session keeps working, 30
minutes by default. [`POST /v1/auth/verify`](api/reference.md#post-v1authverify)
does not change that: it makes the checks a local verifier makes. Nor does
[`POST /v1/auth/me`](api/reference.md#post-v1authme), which makes the same
checks and then reads only the user. An application that needs revocation to
take effect at once cannot get it from access tokens.

A refresh already running when its session is revoked can also still finish,
as if it had finished just before; the refresh token it returns is useless.

### Sessions have no maximum age

Each refresh issues a token with a fresh `REFRESH_TOKEN_TTL`, so a client that
refreshes regularly stays signed in indefinitely: `REFRESH_TOKEN_TTL` limits how
long a session can go unused, not how long it can last.

### A repeated refresh logs the user out

When a refresh token arrives a second time, the service cannot tell the
legitimate client from someone who copied the token, so it revokes the session
([why](architecture/usecases.md#reuse-detection)). A legitimate client sends a
token twice too, when a response is lost and it retries, or when tabs sharing
storage refresh at once. Either way the user signs in again. Clients avoid it
by following the [client obligations](api/reference.md#client-obligations).

## A lone surrogate in a password becomes `U+FFFD`

A body in invalid UTF-8 is rejected, but an escaped lone surrogate such as
`\ud800` is valid UTF-8 on the wire, and the JSON decoder replaces it with
`U+FFFD` without an error. In a password, that replacement is what gets hashed,
so:

- **different passwords can be the same one**: `ab\ud800` and `ab\udfff` log in
  to the same account;
- **the password is longer than sent**: each replacement counts as 3 bytes
  toward the 72-byte limit;
- **the account depends on the replacement**: a decoder that rejected lone
  surrogates would lock that user out.

A client sends one only when its string is not well-formed Unicode, such as a
JavaScript string cut through the middle of an emoji.
