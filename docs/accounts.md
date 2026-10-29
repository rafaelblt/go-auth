# Accounts

Users, their passwords, and the endpoints that take a password:
[registration](#registration), [login](#login) and
[changing the password](#changing-the-password). What a user and a password
are: [Concepts](system.md#concepts). What every endpoint shares, such as the
content type, the request body and the error shape:
[API conventions](api/conventions.md).

## Usernames and passwords

| Field | Rules |
|---|---|
| Username | 3–32 code points; only `a-z`, `0-9`, `.`, `_`, `-`, after lower-casing |
| Password | at least 8 code points, at most 72 bytes; any character |

- A username is lower-cased before it is checked and stored, and never
  trimmed: `Alice` and `alice` are one user, stored as `alice`, and `" alice"`
  breaks the character rule.
- A password is stored and compared exactly as sent: no character is
  forbidden, and nothing is trimmed or normalised.
- The rules are not configurable, and cannot be tightened within v1
  ([why](../CONTRIBUTING.md#compatibility)).

A broken rule is reported with its [validation code](api/conventions.md#validation-errors),
and a length with the unit it was counted in.

**Why:** a username is an identifier, not a display name: its narrow character
set keeps it safe in URLs and logs, and rules out look-alike characters and
bidirectional controls, at the cost of non-Latin scripts. A password has to
work at login exactly as it was typed at registration, through any change of
hashing library, so nothing normalises it. 72 bytes is bcrypt's input limit:
bcrypt ignores what comes after, so two longer passwords sharing their first 72
bytes would match each other. The minimum counts code points because it is
about how much password there is: in bytes, 4 characters of a non-Latin script
would pass where 4 ASCII characters fail.

## Registration

`POST /v1/auth/register` creates a user and its password. It does not log the
user in.

**Request**

```json
{ "username": "alice", "password": "correct-horse" }
```

**Response: `201 Created`**

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

**Errors**

| Status | Code | When |
|---|---|---|
| `409` | `username_already_exists` | The username is taken: ask for another |
| `422` | `validation_failed` | The username or the password breaks a [rule](#usernames-and-passwords) |

A `422` lists every rule broken, in both fields, so one request reports every
problem:

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

- Whether a name is free is checked before the password is hashed, so a taken
  name costs no bcrypt, but only storing the user decides it: of any number of
  registrations of one name sent at once, exactly one succeeds and the others
  get the `409`.
- The user and its password are stored together or not at all, so no account
  is left without a password, holding its username forever.

## Checking credentials

Login and change password check a username and a password the same way:

1. The username and the password are read with the
   [registration rules](#usernames-and-passwords). A value that breaks one is
   rejected at once, like a wrong password: never as a validation error.
2. The user is looked up, then its password.
3. The password is compared with the stored hash.

Every failure gets the same answer, `invalid_credentials`, and the log says
which it was ([reasons](logging.md#outcomes)). What the answer keeps secret is
whether a password is right, not whether its account exists, which
[registration reveals](#registration-reveals-whether-a-username-exists).

**An unknown username takes as long as a wrong password.** When the user or its
password is missing, the password is compared against a dummy hash, made at
startup with the current `BCRYPT_COST`, so the miss costs the same bcrypt
comparison. What is left: a missing user skips one database read, small next to
bcrypt and network jitter, and a change of `BCRYPT_COST` reopens the gap
([Limitations](#login-timing-after-a-bcrypt_cost-change)). A value that breaks a
rule needs no such cover: its timing depends only on the input, and the rules
are public.

**Why:** a bcrypt comparison costs hundreds of milliseconds, so answering a
missing account at once would tell it apart by time.

## Login

`POST /v1/auth/login` [checks the credentials](#checking-credentials), opens a
[session](sessions.md#life-of-a-session), and returns an access token and the
session's first refresh token. The session and its first token are stored
together.

**Request**

```json
{ "username": "alice", "password": "correct-horse" }
```

**Response: `200 OK`**

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

| Field | Meaning |
|---|---|
| `access_token.value` | The [access token](access-tokens.md): send it to your services |
| `refresh_token.value` | The session's refresh token ([format](system.md#the-two-tokens)), for [refreshing](sessions.md#refreshing). Treat it as a password: keep it private, never log it, never put it in a URL |
| `expires_at` | When the token expires: `ACCESS_TOKEN_TTL` or `REFRESH_TOKEN_TTL` after it was issued |
| `expires_in` | Whole seconds from the moment the service issued the token to its expiry, rounded down |

If you cannot trust your own clock, count `expires_in` from when the response
arrives, and refresh a little early: the time on the network is not deducted.

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_credentials` | Any failure of the [credential check](#checking-credentials): tell the user "wrong username or password" |

**No session opened with a password outlives a change of that password.** A
login checks the password before it saves its session. If a change of that
password commits in between, the login does not save the session, never hands
out the token it issued, and answers `401`, as it would have a moment later. If
the login saves first, the change revokes its session with the others. So a
`401` can answer a password that was right when it was sent.

## Changing the password

`POST /v1/auth/change-password` takes the username, the current password and a
new one, replaces the password, and revokes every session of the user, the
caller's included. It returns no tokens: the user logs in again with the new
password.

**Request**

```json
{ "username": "alice", "current_password": "correct-horse", "new_password": "battery-staple" }
```

In order:

1. `new_password` is checked: the [password rules](#usernames-and-passwords),
   and not `current_password`, byte for byte. No older password is
   remembered, so one used before is accepted again. A failure comes before the
   account is looked up, so it says nothing about the account and costs no
   bcrypt.
2. The username and the current password are
   [checked as at login](#checking-credentials).
3. The password is replaced and every session revoked, together: from then on,
   every refresh token issued before the change is rejected. It does not reach
   the access tokens already issued
   ([Access tokens cannot be revoked](access-tokens.md#access-tokens-cannot-be-revoked)).

**Response: `200 OK`**

```json
{ "user": { "id": "0f1c2e5a-7b3d-4c8e-9a1f-2b6d4e8c0a37", "username": "alice" } }
```

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_credentials` | Step 2 failed: tell the user "wrong username or current password" |
| `422` | `validation_failed` | Step 1 failed; `fields` names only `new_password`, with `same_as_current` when it is the current password |

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

Of any number of changes sent at once with the same current password, exactly
one applies; the others get the `401`, since the password they presented is no
longer the user's. A login running at the same moment:
[Login](#login).

**Why every session:** a password is most often changed when someone else may
know it, and a session has [no maximum age](sessions.md#sessions-have-no-maximum-age):
a refresh token obtained with the old password would otherwise keep its holder
signed in for good.

## Limitations

### Registration reveals whether a username exists

Registration answers `username_already_exists` for a taken username, so anyone
can learn whether an account exists, as often as the
[rate limit](api/conventions.md#rate-limiting) lets them, or without limit when
it is off.

### Whoever knows the password can lock the owner out

Changing the password asks for the username and the current password, and
nothing else: no second factor, no confirmation sent elsewhere, no notice to the
owner. The change revokes every session, the owner's included, and the service
offers no way back: no password reset, no account management. Getting the
account back is up to whatever the deployment builds around the service. It
works the other way too: an owner who changes the password signs out whoever
else held it.

### Login timing after a `BCRYPT_COST` change

A stored hash keeps the cost it was made with until its user changes the
password, while the [dummy hash](#checking-credentials) is made with the current
`BCRYPT_COST`. Once the cost changes, the accounts hashed before no longer take
as long as a missing one, and response time shows which of those usernames
exist. Raising the cost also leaves those older hashes as easy to crack offline
as before. To avoid both, choose `BCRYPT_COST` before the first user registers,
and keep it.

### A lone surrogate in a password becomes `U+FFFD`

A body in invalid UTF-8 is [rejected](api/conventions.md#request-bodies), but an
escaped lone surrogate such as `\ud800` is valid UTF-8 on the wire, and the JSON
decoder replaces it with `U+FFFD` without an error. In a password, that
replacement is what gets hashed, so:

- **different passwords can be the same one**: `ab\ud800` and `ab\udfff` log in
  to the same account;
- **the password is longer than sent**: each replacement counts as 3 bytes
  toward the 72-byte limit;
- **the account depends on the replacement**: a decoder that rejected lone
  surrogates would lock that user out.

A client sends one only when its string is not well-formed Unicode, such as a
JavaScript string cut through the middle of an emoji.
