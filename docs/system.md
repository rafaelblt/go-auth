# How go-auth works

What the service is, the concepts it works with, and where each part is
described. Each topic document holds its behaviour, its endpoints, the reasons
behind them and its limitations, without reference to the code; what every
endpoint shares is in [API conventions](api/conventions.md), the settings in
[Configuration](configuration.md), the log lines in [Logging](logging.md), and
how the code does it in the [architecture guides](architecture/overview.md).

## What it is

`go-auth` is a standalone HTTP service that owns a PostgreSQL database and
answers in JSON, so applications in any language can use it. It has two kinds
of callers:

- **client applications**, which sign their users up and in
  ([Accounts](accounts.md)), and keep them signed in with refresh tokens
  ([Sessions](sessions.md));
- **your services**, which receive access tokens with each request and check
  them, either themselves, with the public keys the service publishes, or by
  asking the service ([Access tokens](access-tokens.md)).

It runs as one or more identical instances over the same database, behind a
reverse proxy that terminates TLS ([Deploying](getting-started.md#deploying)).

## Concepts

```
user ──1:0..1── password
 │
 └──1:N── session ──1:N── refresh token ── replaced by ──► refresh token
```

- A **user** is an account: an ID, a UUID, and a username. A user has one
  status, active, and nothing in the service deletes one.
- A **password** belongs to one user, at most one per user, and is stored as a
  bcrypt hash that keeps the cost it was made with ([Accounts](accounts.md)).
- A **session** is one login ([Life of a session](sessions.md#life-of-a-session)).
- A **refresh token** keeps a session going, and an **access token** is what
  the user's requests carry to your services: [The two tokens](#the-two-tokens).
- **Signing keys** sign the access tokens, and their public halves are
  published at `/.well-known/jwks.json`
  ([Signing keys](access-tokens.md#signing-keys)).

### The two tokens

A token that can be checked without calling the service cannot be revoked, and
a token that can be revoked has to be checked by the service. So there are two,
with opposite trade-offs:

| | Access token | Refresh token |
|---|---|---|
| Format | JWT, EdDSA (Ed25519) ([format](access-tokens.md#token-format)) | 32 random bytes, base64url |
| Checked by | any service, with the JWKS, or this service on request | this service only |
| Lifetime | `ACCESS_TOKEN_TTL` | `REFRESH_TOKEN_TTL`, renewed by every [refresh](sessions.md#rotation) |
| Stored | no | as its SHA-256 |
| Used | until it expires | once: every refresh replaces it |
| Revocable | no ([consequences](access-tokens.md#access-tokens-cannot-be-revoked)) | yes, with its session |

The lifetimes and their defaults are in [Configuration](configuration.md).

**Why:** the refresh token is looked up on every use, so it needs no content,
only to be impossible to guess, and storing its digest means a leaked database
gives out no working token.

## Endpoints

| Method | Path | Purpose | Described in |
|---|---|---|---|
| `POST` | `/v1/auth/register` | Create a user | [Registration](accounts.md#registration) |
| `POST` | `/v1/auth/login` | Exchange credentials for a token pair | [Login](accounts.md#login) |
| `POST` | `/v1/auth/change-password` | Replace a password, given the current one, and revoke every session | [Changing the password](accounts.md#changing-the-password) |
| `POST` | `/v1/auth/refresh` | Exchange a refresh token for a new pair | [Refreshing](sessions.md#refreshing) |
| `POST` | `/v1/auth/logout` | End the session of a refresh token | [Logging out](sessions.md#logging-out) |
| `POST` | `/v1/auth/verify` | Check an access token, for a service that does not verify it itself | [Letting go-auth verify](access-tokens.md#letting-go-auth-verify) |
| `POST` | `/v1/auth/me` | Return the user an access token was issued to | [Looking up the user](access-tokens.md#looking-up-the-user) |
| `GET` | `/.well-known/jwks.json` | Public keys for verifying access tokens | [The JWKS](access-tokens.md#the-jwks) |

## Out of scope

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
- TLS or CORS;
- cookies: tokens travel in JSON bodies
  ([where a browser keeps them](sessions.md#tokens-in-a-browser)).

## Limitations

The trade-offs to know before deploying are at the end of each topic, next to
the behaviour that causes them. They are current behaviour, not defects
waiting for a fix:

- [Accounts](accounts.md#limitations);
- [Sessions](sessions.md#limitations);
- [Access tokens](access-tokens.md#limitations);
- [Rate limiting](api/conventions.md#limitations);
- and what a deployment has to provide: [Deploying](getting-started.md#deploying).

## Data

Everything the service knows is in its PostgreSQL database. Every time is
stored and read in UTC.

| What | Kept | Deleted |
|---|---|---|
| Users | ID, username, status | never |
| Passwords | the bcrypt hash, with its cost and salt | never: a change replaces the hash |
| Sessions | when they started and when they were revoked | by the [session cleanup](sessions.md#session-cleanup) |
| Refresh tokens | the SHA-256 of each, its expiry, when it was used and which token it replaced | with their session |
| Signing keys | the private key, [sealed](access-tokens.md#sealed-keys) or not, and when it starts signing | once retired ([Signing keys](access-tokens.md#signing-keys)) |

### Migrations

The schema is versioned by migrations embedded in the binary. At startup the
service applies the pending ones when `AUTO_MIGRATE` is set, and either way
refuses to start unless the schema is exactly the version it was built for, in
either direction. A released migration is never edited; a change is a new one.
How to apply them as an operator: [Migrations](getting-started.md#migrations).

**Why:** a binary running against a schema it was not built for would otherwise
fail one request at a time.
