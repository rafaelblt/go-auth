# Sessions

A session is one login, kept going by refresh tokens. This document covers what
happens to it after [login](accounts.md#login) opens it:
[refreshing](#refreshing), [logging out](#logging-out), where a browser keeps
the tokens, and the [session cleanup](#session-cleanup). What every endpoint
shares: [API conventions](api/conventions.md).

## Life of a session

- A session is **active** from its login until it is **revoked**, by a
  [logout](#logging-out), by [reuse detection](#reuse-detection) or by a
  [password change](accounts.md#changing-the-password). Every cause revokes it
  the same way: the data keeps when it was revoked, not why, and the log says
  why.
- Its **refresh tokens** ([what they are](system.md#the-two-tokens)) form a
  chain: each [refresh](#rotation) spends one and issues its successor.
- It is **dead** once it is revoked or its latest refresh token has expired:
  nothing can refresh it again. The [session cleanup](#session-cleanup)
  deletes it later.

Revoking a session stops its refresh tokens at once, not its access tokens
([Access tokens cannot be revoked](access-tokens.md#access-tokens-cannot-be-revoked)).

## Refreshing

`POST /v1/auth/refresh` exchanges a refresh token for a new access token and a
new refresh token.

**Request**

```json
{ "refresh_token": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01" }
```

In order:

1. A token that is malformed or unknown is rejected.
2. So is a token whose session no longer exists, because the
   [session cleanup](#session-cleanup) deleted it: the token is unknown from
   then on, as in step 1.
3. So is a token whose session was revoked.
4. A token already used goes to [reuse detection](#reuse-detection), and a
   token past its expiry is rejected.
5. Otherwise, [rotation](#rotation).

**Response: `200 OK`**, the same shape as the [login response](accounts.md#login).
**The refresh token you sent is now spent**: replace your stored copy with the
new one.

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_token` | Steps 1 to 4: discard the token and send the user to log in |

One code for every failure, so the response tells someone holding a stolen
token nothing about its state. The log says which step it was
([reasons](logging.md#outcomes)).

### Rotation

Every refresh spends the token presented and returns a new one, which gets a
full `REFRESH_TOKEN_TTL` and records the token it replaced. Spending the old
token and storing the new one happen together, so a failure in between cannot
burn the client's only token.

Of two refreshes with the same token at once, exactly one rotates. The other is
answered as reuse, and revokes the session.

**Why:** a token that stayed valid until it expired would be worth stealing for
that whole time, and a copy could not be told from the original; rotation
limits that, and is what makes reuse detectable at all. Two refreshes arriving
together are what reuse detection describes, and arriving together makes
neither more trustworthy: if the attacker's arrives first, revoking is the only
thing that locks them out.

### Reuse detection

A token that was already used revokes its whole session before the `401`: the
session's current refresh token stops working too, and the user has to log in
again. "Already used" is checked before "expired", so replaying a token that
was spent and has since expired still revokes the session.

If the session cleanup deletes the session while it is being revoked, the token
is answered as unknown, as in step 2. If anything else revokes it first, the
token is answered as one of a revoked session, as in step 3, and only the
request that revoked the session is logged as reuse. A refresh already past
step 3 when its session is revoked can still finish, as if it had finished just
before, and the refresh token it returns belongs to the revoked session. The
chain of tokens keeps the rotation history of a compromised session.

**Why:** a refresh token is used once, so a second presentation means two
parties hold it, the client and someone who copied it, and the service cannot
tell which one is calling. Revoking locks out both: the attacker for good, the
user until they log in again. Serving the request instead would let an attacker
who replayed a token keep a rotating credential indefinitely. Checking
"expired" first would let someone holding an old stolen token wait for it to
expire, replay it, and never trip the alarm. The legitimate client pays when it
sends a token twice itself
([Limitations](#a-repeated-refresh-logs-the-user-out)).

### Client obligations

1. Store the refresh token privately; losing it costs only a new login. In a
   browser, where to keep it is a trade-off
   ([Tokens in a browser](#tokens-in-a-browser)).
2. After a successful refresh, replace it with the new one at once.
3. Never retry a refresh with the old token: if the response was lost, the
   retry trips reuse detection and logs the user out. Log in again instead.
4. Never run two refreshes with the same token at once: across tabs, threads
   or instances sharing storage, only one may be in flight.
5. On a `401`, discard the token and send the user to log in.

A `429` is the exception to point 3: nothing ran
([Rate limiting](api/conventions.md#rate-limiting)), so wait for `Retry-After`
and send the same token.

## Logging out

`POST /v1/auth/logout` ends the session of a refresh token: none of the
session's refresh tokens can be used again, and the user's other sessions stay
active. It takes the refresh token from the last login or refresh.

**Request**

```json
{ "refresh_token": "kZ8m2Q1nR7yTxV3bC0dEfGhIjKlMnOpQrStUvWxYz01" }
```

In order:

1. A token that is malformed or unknown is rejected, and so is a token whose
   session the session cleanup deleted, as in steps 1 and 2 of
   [refreshing](#refreshing).
2. A session already revoked is left as it is.
3. Otherwise the session is revoked, whatever state the token is in: current,
   already used, or expired. A token already used is [reuse](#reuse-detection),
   and is logged as such ([Outcomes](logging.md#outcomes)).

**Response: `204 No Content`**, for steps 2 and 3, with no body and no
`Content-Type`. So a logout can be sent again, and its answer says nothing
about the state of the token, only that the service knows it. It does not reach
the access tokens already issued
([Access tokens cannot be revoked](access-tokens.md#access-tokens-cannot-be-revoked)).

**Errors**

| Status | Code | When |
|---|---|---|
| `401` | `invalid_token` | Step 1, a missing field included: nothing is left to end, so discard the token |

A refresh with the same token that is already past step 3 of refreshing when
the logout revokes the session can still finish, as after any revocation: the
refresh token it returns belongs to the revoked session. Avoiding that is the
client's part ([obligations](#client-obligations-on-logout)). Of the requests
revoking the session at once, the first keeps its time, and a logout that comes
after it is answered as in step 2. If the session cleanup deletes the session
while a logout runs, the token is answered as unknown, as it would be a moment
later.

**Why:** it takes the refresh token because the access token names no session,
passes through every service that checks it, and may have expired by the time
the user logs out; whoever holds a refresh token could already end its session
through reuse detection. A spent token still ends the session because, when an
attacker refreshes first, the client is left holding a spent token, and its
logout has to end the attacker's chain of tokens. An unknown token is a `401`,
where [RFC 7009](https://www.rfc-editor.org/rfc/rfc7009) would answer a
revocation with success either way, because a missing or misspelled field is
read as an empty token ([Request bodies](api/conventions.md#request-bodies)),
and a success there would hide a client that never logs anyone out.

### Client obligations on logout

1. Log the user out locally at once, whatever happens next: stop using both
   tokens. Keep the refresh token only until a logout with it answers `204` or
   `401`, either of which means there is nothing left to end.
2. Send the logout again, with the same token, after a `5xx`, a `429` (once
   `Retry-After` has passed) or a network error: until a logout succeeds, the
   session is alive. Unlike a refresh, which must never be repeated, a logout
   can be: one that went through answers `204` again.
3. Log out under the lock that keeps two refreshes from running at once
   ([point 4](#client-obligations)), and discard the answer of a refresh that
   arrives after the logout was sent: its access token would keep the user
   signed in until it expires. A logout that races a refresh with the same
   token can also be logged as reuse.

## Tokens in a browser

`go-auth` returns tokens in JSON bodies and sets no cookies, so where they are
kept is up to the client application. A page that receives them can keep them
only where its own scripts can read them, in memory or in web storage, and a
script injected into the page can read them there too. The refresh token is
the one to protect: each refresh issues another, so a stolen copy can keep the
session going until the user's own client refreshes again and trips
[reuse detection](#reuse-detection).

The ways to keep them, from the most protected to the most convenient:

- **On the application's server, which is the recommended way.** A
  backend-for-frontend calls `go-auth`, holds both tokens, and gives the
  browser a cookie of its own, `HttpOnly`, `Secure` and `SameSite`. No script
  can read that cookie, so an injected one can act as the user only from the
  open page, and cannot take the session elsewhere. `go-auth` then need not be
  reachable from browsers, and keeping refreshes one at a time
  ([client obligations](#client-obligations)) is simpler on a server than
  across tabs. Its users share one rate-limit allowance unless it forwards
  their addresses ([Which address counts](api/conventions.md#which-address-counts)).
- **In the page's memory**, for a page that calls `go-auth` itself, which it
  can only from the same origin ([why](api/conventions.md#content-type)). Only
  that page can reach the tokens, and only while it is open; the cost is a new
  login on every page load and in every tab.
- **In web storage, such as `localStorage`.** The user stays signed in across
  page loads and tabs, but any script that runs on the origin can read the
  refresh token for as long as it is valid. An application that chooses it for
  that convenience can make injection harder with a strict Content Security
  Policy.

## Session cleanup

Every instance periodically ([how often](architecture/startup.md#background-tasks))
deletes the sessions dead for longer than
[`SESSION_RETENTION`](configuration.md#session_retention), with their refresh
tokens. It never runs while the retention is `off`.

- **Whole sessions, never a token on its own.**
- **By the expiry each token was issued with**, never by the current
  `REFRESH_TOKEN_TTL`, which can change between deploys.
- **A bounded number per run**, so a backlog, such as a deployment that ran
  without the cleanup has, takes as many runs as it needs. Instances running it
  at the same time share the work rather than wait on each other.

A refresh that presents a token of a session at the moment it is deleted, a
token dead for longer than the retention and refused either way, is answered as
unknown, as it would be a moment later.

**Why:** without it, the tables grow for as long as the deployment runs. It
never deletes a token on its own because a spent token of a live session is
what reuse detection looks for: deleted, its replay would be answered as unknown
instead of revoking the session. A dead session's tokens can no longer be
refreshed, so nothing is lost with them.

## Limitations

### Sessions have no maximum age

Each refresh issues a token with a full `REFRESH_TOKEN_TTL`
([Rotation](#rotation)), so a client that refreshes regularly stays signed in
indefinitely: `REFRESH_TOKEN_TTL` limits how long a session can go unused, not
how long it can last.

### A repeated refresh logs the user out

A legitimate client sends a token twice too, when a response is lost and it
retries, or when tabs sharing storage refresh at once, and
[reuse detection](#reuse-detection) cannot tell it from someone who copied the
token: the user signs in again. Clients avoid it by following the
[client obligations](#client-obligations).

### A session in use keeps every refresh token

Every login stores a session and a refresh token, and every refresh a refresh
token. The cleanup deletes only [whole dead sessions](#session-cleanup), so a
session that stays in use keeps every token it was issued, and, with
[no maximum age](#sessions-have-no-maximum-age), keeps adding them: a client
that refreshes every 30 minutes adds 48 rows a day for as long as it stays
signed in. With `SESSION_RETENTION` set to `off`, nothing is deleted at all.

### The session cleanup can fall behind

The cleanup's first run on an instance comes one interval after it starts, and
each run deletes a bounded number of sessions
([both](architecture/startup.md#background-tasks)). As a result:

- **an instance that lives less than an interval never runs it.** When every
  instance is replaced sooner, by frequent deploys, autoscaling or a platform
  that stops idle instances, nothing is deleted, as with `SESSION_RETENTION`
  set to `off`.
- **a deployment that opens sessions faster than its instances delete them
  falls behind**, and its dead sessions pile up until the rate drops. Its runs
  then keep stopping at their limit, which the line each one writes says
  ([Background tasks](logging.md#background-tasks)).
