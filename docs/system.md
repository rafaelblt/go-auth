# How go-auth works

What the service does, why, and what it does not do, without reference to the
code. Each fact about its behaviour is stated here once: the other documents
link to it. What a client sends and receives is in the
[API reference](api/reference.md), the settings in
[Configuration](configuration.md), the log lines in [Logging](logging.md), and
how the code does it in the [architecture guides](architecture/overview.md).

## What it is

`go-auth` is a standalone HTTP service that owns a PostgreSQL database and
answers in JSON, so applications in any language can use it. It has two kinds
of callers:

- **client applications**, which register users, log them in and out, keep
  them signed in with refresh tokens, change passwords, and ask whose access
  token they hold;
- **your services**, which receive access tokens with each request and check
  them, either themselves, with the public keys the service publishes, or by
  asking the service.

Being a service rather than a library costs a network round trip on every call.
That is why access tokens can be checked without calling it: the service is
involved in logins and refreshes, which are rare, and not in the requests your
application serves.

It runs as one or more identical instances over the same database, behind a
reverse proxy that terminates TLS ([Limitations](#limitations)).

## Concepts

```
user ──1:0..1── password
 │
 └──1:N── session ──1:N── refresh token ── replaced by ──► refresh token
```

- A **user** is an account: an ID, a UUID, and a username. A user has one
  status, active, and nothing in the service deletes one.
- A **password** belongs to one user, at most one per user, and is stored as a
  bcrypt hash that keeps the cost it was made with. It is separate from the
  user so that another kind of credential could be added without changing what
  a user is.
- A **session** is one login. It is active until it is revoked, by a
  [logout](#logging-out), by [reuse detection](#reuse-detection) or by a
  [password change](#changing-the-password). It is **dead** once it is revoked
  or its latest refresh token has expired: nothing can refresh it again.
- A **refresh token** keeps a session going. The client holds it, and the
  service stores only its SHA-256. Each one is used once: a refresh spends it
  and issues its successor, so a session's tokens form a chain.
- An **access token** is what the user's requests carry to your services: a
  signed JWT naming the user and its expiry
  ([format](api/token-verification.md#token-format)).
- **Signing keys** sign the access tokens. They are stored in the database,
  shared by every instance, and replaced on a schedule; their public halves are
  published at `/.well-known/jwks.json` ([Signing keys](#signing-keys)).

### The two tokens

A token that can be checked without calling the service cannot be revoked, and
a token that can be revoked has to be checked by the service. So there are two,
with opposite trade-offs:

| | Access token | Refresh token |
|---|---|---|
| Format | JWT, EdDSA (Ed25519) | 32 random bytes, base64url |
| Checked by | any service, with the JWKS, or this service on request | this service only |
| Lifetime | `ACCESS_TOKEN_TTL`, 30 minutes by default | `REFRESH_TOKEN_TTL`, 7 days by default |
| Stored | no | as its SHA-256 |
| Used | until it expires | once: every refresh replaces it |
| Revocable | no | yes, with its session |

The refresh token is where control lies, so `ACCESS_TOKEN_TTL` is how long a
revoked session can still present a valid access token
([Access tokens cannot be revoked](#access-tokens-cannot-be-revoked)).

The refresh token is random bytes rather than a JWT: it is looked up on every
use, so it needs no content, only to be impossible to guess, and a JWT would
invite clients to read claims from it. A leaked database gives out digests, not
working tokens. A slow hash would add nothing: there is nothing to guess in 32
random bytes, and bcrypt would cost every refresh hundreds of milliseconds.

## Processes

Every failure below answers one generic code, and the log says which failure it
was, as a reason ([Deliberately vague answers](#deliberately-vague-answers));
the reasons are listed in [Logging](logging.md#outcomes).

### Registration

Creates a user and its password. It does not log the user in.

- The username and the password are checked against the
  [rules](#usernames-and-passwords), and every rule they break is reported in
  one `422`.
- A taken username is a `409`. Whether a name is free is checked before the
  password is hashed, so a taken name costs no bcrypt, but only storing the user
  decides it: of any number of registrations of one name sent at once, exactly
  one succeeds and the others get the `409`.
- The user and its password are stored together or not at all, so no account
  is left without a password, holding its username forever.

### Usernames and passwords

The rules themselves are in the [API reference](api/reference.md#post-v1authregister).
Why they are what they are:

- **A username is an identifier, not a display name.** Its narrow character set
  keeps it safe in URLs and logs, and rules out look-alike characters and
  bidirectional controls, so two usernames cannot look the same while differing
  in bytes. The cost is that non-Latin scripts cannot be used.
- **Usernames are lower-cased, never trimmed.** `JOAO` and `joao` are one
  identity, stored as `joao`; `" alice"` fails the character check rather than
  registering a name the user did not type.
- **A password is stored exactly as typed.** Whatever was typed at registration
  must work at login, so no character is forbidden and nothing is trimmed or
  normalised: normalising would mean normalising the same way forever, through
  any change of hashing library.
- **At most 72 bytes**, bcrypt's input limit: bcrypt ignores what comes after,
  so two longer passwords sharing their first 72 bytes would match each other.
- **At least 8 code points**, the minimum chosen for this service. It counts
  code points because it is about how much password there is: in bytes, 4
  characters of a non-Latin script would pass where 4 ASCII characters fail.
  There are no composition rules: the minimum length is the whole policy.
- **Not configurable.** A stored username is rebuilt with these rules when it
  is read, and login checks the credentials against them, so tightening one
  would lock out the users stored before
  ([Compatibility](../CONTRIBUTING.md#compatibility)).

A length is always reported with its unit, `code_point` or `byte`, because
clients count differently: JavaScript's `length` counts UTF-16 units, Go's
`len` bytes, and a user characters.

### Login

Checks a username and a password, opens a session, and returns an access token
and the session's first refresh token, which expires `REFRESH_TOKEN_TTL` later.
The session and its first token are stored together.

Every failure is `401 invalid_credentials`: an unknown username, a user without
a password, a wrong password, and a username or password that breaks the
registration rules, which never gets a `422`.

**An unknown username takes as long as a wrong password.** Checking a password
costs a bcrypt comparison, hundreds of milliseconds, so answering a missing
account at once would tell it apart by time. A missing account is therefore
compared against a dummy hash, made at startup with the current `BCRYPT_COST`.
What is left: a missing user skips one database read, small next to bcrypt and
network jitter; a change of `BCRYPT_COST` reopens the gap for the accounts
hashed before it ([Limitations](#login-timing-after-a-bcrypt_cost-change)); and
registration says whether a username is taken anyway
([Limitations](#registration-reveals-whether-a-username-exists)).

**No session opened with a password outlives a change of that password.** A
login checks the password before it saves its session. If a change of that
password commits in between, the login does not save the session, never hands
out the token it issued, and answers `401`, as it would have a moment later. If
the login saves first, the change revokes its session with the others. So a
`401` can answer a password that was right when it was sent.

### Changing the password

Takes the username, the current password and a new one; replaces the password
and revokes every session of the user, the caller's included. It returns no
tokens: the user logs in again with the new password.

1. The new password is checked first: the registration rules, and not the
   current password byte for byte (no older password is remembered). A failure
   is a `422` naming only `new_password`. It comes before the account is
   looked up, so it says nothing about the account and costs no bcrypt.
2. The username and the current password are checked [as at login](#login),
   with the same `401 invalid_credentials` and the same reasons.
3. The password is replaced and every session revoked, together.

Every session goes because a password is most often changed when someone else
may know it, and a session has no maximum age: a refresh token obtained with the
old password would otherwise keep its holder signed in for good. The access
tokens already issued keep working until they expire
([Limitations](#access-tokens-cannot-be-revoked)).

Of any number of changes sent at once with the same current password, exactly
one applies; the others get `401`, since the password they presented is no
longer the user's.

### Refreshing

Exchanges a refresh token for a new access token and a new refresh token. In
order:

1. A token that is malformed or unknown is rejected.
2. So is a token whose session no longer exists, because the
   [session cleanup](#session-cleanup) deleted it: the token is unknown from
   then on, as in step 1.
3. So is a token whose session was revoked.
4. A token already used goes to [reuse detection](#reuse-detection), and a
   token past its expiry is rejected.
5. Otherwise, [rotation](#rotation).

Every failure is `401 invalid_token`, so the response tells someone holding a
stolen token nothing about its state.

#### Rotation

Every refresh spends the token presented and returns a new one, which gets a
full `REFRESH_TOKEN_TTL` and records the token it replaced. A token that stayed
valid until it expired would be worth stealing for that whole time, and a copy
could not be told from the original; rotation limits that, and is what makes
reuse detectable at all. Spending the old token and storing the new one happen
together, so a failure in between cannot burn the client's only token.

Of two refreshes with the same token at once, exactly one rotates. The other is
answered as reuse, and revokes the session: two parties presenting one token
are what reuse detection describes, and arriving together makes neither more
trustworthy. If the attacker's arrives first, revoking is the only thing that
locks them out.

#### Reuse detection

A refresh token is used once, so a second presentation means two parties hold
it: the client, and someone who copied it. The service cannot tell which one is
calling, so it revokes the whole session, which locks out both: the attacker for
good, the user until they log in again. Serving the request instead would let an
attacker who replayed a token keep a rotating credential indefinitely.

"Already used" is checked before "expired", so replaying a token that was spent
and has since expired still revokes the session. The other way round, someone
holding an old stolen token could wait for it to expire, replay it, and never
trip the alarm. If the session cleanup deletes the session while it is being
revoked, the token is answered as unknown, as in step 2. If a logout, a
password change or another refresh revokes it first, the token is answered as
one of a revoked session, as in step 3, and only the request that revoked the
session is logged as reuse.

The legitimate client pays when it sends a token twice itself, after a lost
response or from two tabs ([Limitations](#a-repeated-refresh-logs-the-user-out)).
The chain of tokens keeps the rotation history of a compromised session. A
refresh already past step 3 when its session is revoked can still finish, as if
it had finished just before, and the refresh token it returns belongs to the
revoked session.

### Logging out

Ends the session of a refresh token: none of the session's refresh tokens can
be used again, and the user's other sessions stay active. It takes the refresh
token rather than the access token, which names no session, passes through
every service that checks it, and may have expired by the time the user logs
out. Whoever holds a refresh token could already end its session through
[reuse detection](#reuse-detection), so logging out gives no one a power they
lacked. In order:

1. A token that is malformed or unknown is rejected, and so is a token whose
   session the [session cleanup](#session-cleanup) deleted, as in steps 1 and 2
   of [Refreshing](#refreshing): `401 invalid_token`.
2. A session already revoked is left as it is.
3. Otherwise the session is revoked, whatever state the token is in: current,
   already used, or expired.

Steps 2 and 3 answer `204`, with no body. So a logout can be sent again, and
its answer says nothing about the state of the token.

- **A spent token still ends the session.** When an attacker refreshes first,
  the client is left holding a spent token, and its logout has to end the
  attacker's chain of tokens. That is reuse, and the log records it as it does
  at refresh ([Outcomes](logging.md#outcomes)).
- **An unknown token is a `401`**, where RFC 7009 would answer a revocation
  with success either way: a missing or misspelled field is read as an empty
  token ([Request bodies](api/reference.md#request-bodies)), and a success there
  would hide a client that never logs anyone out.
- **The access tokens already issued keep working** until they expire, for up
  to `ACCESS_TOKEN_TTL` ([Limitations](#access-tokens-cannot-be-revoked)).
- **The session is revoked as any other is**: the data keeps when it was
  revoked, not why ([Data](#data)), and the log says it was a logout, or that
  the logout found it already revoked.

A refresh with the same token that is already past step 3 when the logout
revokes the session can still finish, as after any revocation: the refresh
token it returns belongs to the revoked session, and its access token works
until it expires. Avoiding that is the client's part
([obligations](api/reference.md#client-obligations-on-logout)). Of logouts,
reuse detection and password changes running alongside, the first to revoke
the session keeps its time, and a logout that comes after it is answered as in
step 2. If the session cleanup deletes the session while a logout runs, the
token is answered as unknown, as it would be a moment later.

### Checking an access token

`POST /v1/auth/verify` checks an access token for a service that would rather
not check it itself, and answers with the user ID and the expiry.

A token passes when it passes the
[checks a correct verifier makes](api/token-verification.md#what-to-verify) and,
when it carries an `nbf`, that time has come. Every failure, a malformed token
included, is `401 invalid_token`; the log tells an expired token from the
others.

It reads no database, so a token passes here exactly when it would pass a
correct verifier elsewhere, and a service can switch between the two. An access
token names a user, not a session, so the tokens of a revoked session pass until
their `exp`. It is never rate limited: it reads no database, no number of tries
finds a signature that passes, and the services that delegate checking call it
on every request they serve, usually from one address.

### Looking up the user

`POST /v1/auth/me` returns the user an access token was issued to, for a client
that needs more than the user ID the token carries, such as the username to
show.

It checks the token [as verify does](#checking-an-access-token), then reads the
user. A token whose user is no longer stored is `401 invalid_token` too. Nothing
in the service deletes a user, so that takes a change made outside it, such as a
row deleted by hand or a database restored from an older backup.

Like verify, it reads nothing about the token's session, so the tokens of a
revoked session still get their user until they expire. Unlike verify, it is
rate limited: every token that passes costs a database read, and it serves the
clients that hold the tokens, which ask when they need to, not on every request
a service serves.

### Session cleanup

Every instance periodically ([how often](architecture/startup.md#background-tasks))
deletes the sessions dead for longer than
[`SESSION_RETENTION`](configuration.md#session_retention), 30 days by default,
with their refresh tokens, so that the tables stop growing for as long as the
deployment runs. It is on by default, and never runs while the retention is
`off`.

- **Whole sessions, never a token on its own.** A spent token of a live session
  is what reuse detection looks for: deleted, its replay would be answered as
  unknown instead of revoking the session. A dead session's tokens can no longer
  be refreshed, so nothing is lost with them.
- **By the expiry each token was issued with**, never by the current
  `REFRESH_TOKEN_TTL`, which can change between deploys.
- **A bounded number per run**, each in a short statement, so a backlog, such
  as a deployment that ran without the cleanup has, takes as many runs as it
  needs rather than one long transaction. Instances deleting at the same time
  share the work rather than wait on each other.

A refresh that presents a token of a session at the moment it is deleted, a
token dead for longer than the retention and refused either way, is answered
as unknown, as it would be a moment later.

### Signing keys

Access tokens are signed with Ed25519: the services that check them hold the
public keys, and with HMAC every one of them would hold the signing secret and
could mint tokens. Ed25519 keys and signatures are small, and the JWKS format
supports them as standard.

The keys are rows in the database every instance shares, so they survive a
restart, and every instance signs and publishes the same ones. Each key goes
through three windows:

- it is added, and published, **a day** before it starts signing, so by then
  every instance has loaded it, and so has a verifier that fetches the JWKS
  [often enough](api/token-verification.md#fetching-and-caching-keys). The first
  key is the exception: it signs at once, since nothing else can;
- it signs for **7 days**, until the next key starts;
- it stays published for **25 hours** more, an hour longer than the longest
  access token the service can issue, and is then deleted, so a copy of the
  database holds only the keys of the last 9 days or so.

So a key is published for about **9 days**, and the JWKS holds one or two keys.
Each instance re-reads the keys every few minutes
([interval](architecture/startup.md#background-tasks)), and goes by its own
clock to decide which key signs: they all switch at the same moment, give or
take their clock skew, with no message between them. A key stops signing only
once its successor exists, so a rotation that fails leaves the current key
signing, never none.

**Sealed keys.** When [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key)
is set, every key added is stored encrypted, and opens only in its own row. A
key stored before it was set still loads, so setting it needs nothing else: the
old keys age out within about 9 days. Sealing protects a copy of the database,
not the database itself: whoever can write the table can store a key in the
clear, as they could replace a password hash. A key cannot be withdrawn while
the service runs; withdrawing one means
[replacing them all](configuration.md#replacing-the-signing-keys).

## Security

### Deliberately vague answers

Login, change password, refresh, logout, verify and me answer every failure of
a kind with one code, and say which failure it was only in the log, as the
request's `reason` ([Outcomes](logging.md#outcomes)), so nothing is lost for
whoever runs the service.

- **Login and change password** answer a wrong password, an unknown username, a
  user without a password and a malformed value alike, in about the same time
  ([Login](#login)). What they keep secret is whether a password is right, not
  whether its account exists, which registration reveals.
- **Refresh** tells someone holding a stolen token nothing about its state.
- **Logout** answers every token the service knows alike, whatever its state,
  so it tells nothing about the state either: only, unlike refresh, whether the
  service knows the token ([why](#logging-out)).
- **Verify and me** give a service, or a client, one thing to do with any
  failure: turn the request away, or get a new token.

The `message` of an error is set by its status, never by its cause: every `401`
reads the same ([Error model](api/errors.md)).

### Requests from browsers

A browser sends a cross-site `POST` without asking the site first when its body
is a form, URL-encoded or `text/plain`; for any other type, `application/json`
included, it first sends a CORS preflight, which `go-auth` never grants: it sends
no CORS headers. So every `POST` requires `Content-Type: application/json`, and
anything else is a `415` before the body is read. Without it, a page on any site
could send logins and registrations through its visitors' browsers, each from
its own address, past any per-address limit. A reverse proxy that answers CORS
preflights for these endpoints, for origins the operator does not control,
undoes this.

### Rate limiting

Off unless [`RATE_LIMIT`](configuration.md#rate_limit) chooses a level, which
sets the allowance of each limited endpoint. The allowances are per client
address and per endpoint: a login does not spend a client's refreshes.

- **The limit is decided before the body is read**, so a refused request read
  and ran nothing: a `429` never spends a refresh token, and the same request
  can be sent again after `Retry-After`.
- **Requests answered `415` are never counted**: they are the ones a browser
  sends cross-site without a preflight, and counting them would let any page
  spend its visitors' allowance.
- **Every other request counts**, whatever its answer: the decision is taken
  before it is known.
- **A failure of the limiter is a `500`**, and the request does not run: failing
  open would remove the limit without anyone noticing.

#### Which address counts

The client address is the TCP peer, unless the peer is one of the
[`TRUSTED_PROXIES`](configuration.md#trusted_proxies). Then `X-Forwarded-For` is
read from the right, skipping each trusted proxy, and the client is the first
address that is not one, or the leftmost when all are. An empty entry, or one
that is not a bare address (`ip:port` included), stops the walk at the last good
address. Reading from the right is what keeps a client from choosing its
address: the leftmost entries are whatever it sent, and only those a trusted
proxy appended can be believed.

Addresses are normalised alike: an IPv4-mapped IPv6 address counts as IPv4, and
a zone is dropped. An IPv6 client counts as its /64, since one host normally
holds a whole /64 and could otherwise rotate through it. The address counted is
logged with each refusal ([Outcomes](logging.md#outcomes)); the request lines
keep the peer.

### Properties

| Property | Holds |
|---|---|
| Passwords stored as bcrypt hashes, at a configurable cost | yes |
| Refresh tokens stored only as their SHA-256 | yes |
| Refresh tokens single use, replaced on every refresh | yes |
| Reusing a refresh token revokes its session | yes |
| Changing the password revokes every session of the user | yes |
| Login and change password answer every credential failure alike | yes ([why](#deliberately-vague-answers)) |
| A missing account takes as long as a wrong password | yes, until `BCRYPT_COST` changes ([why](#login-timing-after-a-bcrypt_cost-change)) |
| Whether a username exists is secret | **no**: registration says it ([why](#registration-reveals-whether-a-username-exists)) |
| Access tokens can be revoked | **no** ([why](#access-tokens-cannot-be-revoked)) |
| Signing keys survive a restart, and rotate without breaking issued tokens | yes |
| Signing keys encrypted at rest | only with [`SIGNING_KEY_ENCRYPTION_KEY`](configuration.md#signing_key_encryption_key) |
| Rate limiting | only with [`RATE_LIMIT`](configuration.md#rate_limit) |
| Credentials, tokens and keys kept out of the logs | yes ([what is never logged](logging.md#what-is-never-logged)) |
| Maximum session age | **no** ([why](#sessions-have-no-maximum-age)) |

## Data

Everything the service knows is in its PostgreSQL database, in five tables:
`users`, `passwords`, `sessions`, `refresh_tokens` and `signing_keys`. Every time
is stored and read in UTC.

| What | Kept | Deleted |
|---|---|---|
| Users | ID, username, status | never |
| Passwords | the bcrypt hash, with its cost and salt | never: a change replaces the hash |
| Sessions | when they started and when they were revoked | by the [session cleanup](#session-cleanup) |
| Refresh tokens | the SHA-256 of each, its expiry, when it was used and which token it replaced | with their session |
| Signing keys | the private key, sealed or not, and when it starts signing | once retired ([Signing keys](#signing-keys)) |

The schema is versioned by migrations embedded in the binary. At startup the
service applies the pending ones when `AUTO_MIGRATE` is set, and either way
refuses to start unless the schema is exactly the version it was built for, in
either direction: a binary running against a schema it was not built for would
otherwise fail one request at a time. A released migration is never edited; a
change is a new one. How to apply them as an operator:
[Migrations](getting-started.md#migrations).

## Limitations

What `go-auth` does not do, and the trade-offs to know before deploying it.
Everything here is current behaviour, not a defect waiting for a fix.

### Out of scope

`go-auth` checks a username and password and issues tokens. It does not
provide:

- email, OAuth or passwordless login, or password reset;
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

### Deployment

#### Run it behind a reverse proxy

`go-auth` serves plain HTTP, and its rate limiting is off unless `RATE_LIMIT`
turns it on. Put it behind a reverse proxy that:

- terminates TLS;
- rate limits the endpoints `RATE_LIMIT` would limit
  ([which](configuration.md#rate_limit)) while it is `off`: nothing else limits
  password guessing then, beyond the cost of a bcrypt comparison;
- exposes only the service's endpoints ([API reference](api/reference.md)).

With the service's own rate limiting on, set `TRUSTED_PROXIES` to the proxy, or
every client is counted as the proxy. A proxy that answers CORS preflights for
these endpoints undoes what the `415` protects
([why](#requests-from-browsers)).

#### Rate limiting is per address and per process

- **Per address, not per account.** Guesses against one account spread over
  many addresses are not limited, whether the proxy or the service limits them.
- **An address can be many users.** Clients behind one NAT share an allowance,
  and so does an IPv6 /64; `relaxed` exists for that. The users of an
  application server that calls the service for them share its allowance too,
  unless it forwards their addresses.
- **In memory.** A restart restores every allowance, and each instance counts
  on its own: N instances allow N times the limit.
- **Off by default.** `off` is deprecated, and v2 removes it.

#### Signing keys are stored in the database

- **They are unencrypted unless `SIGNING_KEY_ENCRYPTION_KEY` is set.** Without
  it, whoever reads the database, or a backup of it, can sign access tokens that
  every verifier accepts while that key is published
  ([how long](#signing-keys)). Protect the database and its backups like the
  keys they hold.
- **The encryption key cannot be changed in place.** Changing it means
  [replacing every signing key](configuration.md#replacing-the-signing-keys),
  which invalidates the access tokens already issued.
- **Instances go by their own clocks** to decide which key signs and when an old
  one is dropped. The windows absorb up to an hour of skew, so keep the clocks
  synchronised.

#### A session in use keeps every refresh token

Every login stores a session and a refresh token, and every refresh a refresh
token. The session cleanup deletes a session, with all its tokens, once it has
been dead for `SESSION_RETENTION`, and never a token on its own
([why](#session-cleanup)). A session that stays in use therefore keeps every
token it was issued, and, with [no maximum age](#sessions-have-no-maximum-age),
keeps adding them: a client that refreshes every 30 minutes adds 48 rows a day
for as long as it stays signed in.

With `SESSION_RETENTION` set to `off`, nothing is deleted, and both tables grow
for as long as the deployment runs.

#### The session cleanup can fall behind

The cleanup's first run on an instance comes one interval after it starts, and
each run deletes a bounded number of sessions
([both](architecture/startup.md#background-tasks)). As a result:

- **an instance that lives less than an interval never runs it.** When every
  instance is replaced sooner, by frequent deploys, autoscaling or a platform
  that stops idle instances, nothing is deleted, and the tables grow as they do
  with `SESSION_RETENTION` set to `off`.
- **a deployment that opens sessions faster than its instances delete them
  falls behind**, and its dead sessions pile up until the rate drops. Nothing in
  the logs shows it: the line a run writes says how long it took, not how many
  sessions it deleted ([Background tasks](logging.md#background-tasks)).

### Accounts

#### Registration reveals whether a username exists

Registration answers `409 username_already_exists` for a taken username, so
anyone can learn whether an account exists, and, unless rate limiting is on, as
often as they like. Login and change password do not tell an unknown username
from a wrong password, so what they keep secret is whether a password is right,
not whether its account exists.

#### Whoever knows the password can lock the owner out

Changing the password asks for the username and the current password, and
nothing else: no second factor, no confirmation sent elsewhere, no notice to the
owner. The change revokes every session, the owner's included, and the service
offers no way back: no password reset, no account management. Getting the
account back is up to whatever the deployment builds around the service. It
works the other way too: an owner who changes the password signs out whoever
else held it.

#### Login timing after a `BCRYPT_COST` change

A missing account is compared against a dummy hash made with the current
`BCRYPT_COST`, while a stored hash keeps the cost it was made with until its
user changes the password. Once `BCRYPT_COST` changes, the accounts hashed
before no longer take as long as a missing one, and response time shows which of
those usernames exist. Raising the cost also leaves those older hashes as easy
to crack offline as before. To avoid both, choose `BCRYPT_COST` before the first
user registers, and keep it.

### Sessions and tokens

#### Access tokens cannot be revoked

Revoking a session, by a logout or any other way, stops its refresh token at
once, but the access tokens already issued stay valid until they expire: they
hold everything needed to check them, not even their session.
`ACCESS_TOKEN_TTL` is therefore how long a revoked session keeps working, 30
minutes by default. Verify and me do not change that: they make the checks a
local verifier makes, and me then reads only the user. An application that
needs revocation to take effect at once cannot get it from access tokens.

#### Sessions have no maximum age

Each refresh issues a token with a fresh `REFRESH_TOKEN_TTL`, so a client that
refreshes regularly stays signed in indefinitely: `REFRESH_TOKEN_TTL` limits how
long a session can go unused, not how long it can last.

#### A repeated refresh logs the user out

When a refresh token arrives a second time, the service cannot tell the
legitimate client from someone who copied the token, so it revokes the session
([why](#reuse-detection)). A legitimate client sends a token twice too, when a
response is lost and it retries, or when tabs sharing storage refresh at once.
Either way the user signs in again. Clients avoid it by following the
[client obligations](api/reference.md#client-obligations).

### A lone surrogate in a password becomes `U+FFFD`

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
