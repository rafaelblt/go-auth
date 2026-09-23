# Refresh

`internal/usecase/refresh`

Exchanges a refresh token for a new pair. The most involved flow, because it
carries reuse detection.

```
Input{RefreshToken}
  │
  ├─ 1. ParseRefreshTokenSecret → SHA-256 → FindByHash
  │       └─ malformed / not found → ErrTokenInvalid → 401
  │
  ├─ 2. Load the token's Session
  │       └─ missing → unexpected error → 500
  │
  ├─ 3. Session revoked? → ErrSessionRevoked → 401
  │
  └─ 4. token.Use(now)
          │
          ├─ ok               → rotate (below)
          ├─ already used     → reuse detected (below)
          ├─ expired          → ErrTokenExpired → 401
          └─ anything else    → 500
```

The four `401` paths share one response, `INVALID_TOKEN`. Each carries a
different `reason` (`invalid token`, `token expired`, `token already used`,
`session revoked`), which never reaches the client. Someone holding a stolen
token learns nothing about its state, and in particular not whether reuse
detection has already fired.

## Rotation

Every refresh replaces the token: the one presented is marked used, and a new
one takes its place. A token that stayed valid until it expired would be worth
stealing for that whole time, and a copy would be impossible to tell from the
original. Rotation limits how long any one token is worth anything, and it is
what makes reuse detectable at all.

```
  ├─ Issue a new access token
  ├─ Create RefreshToken{parent: used.ID, expires: now + REFRESH_TOKEN_TTL}
  │     └─ generates the new secret
  │
  └─ One transaction:
       ├─ MarkUsed the token presented (used_at = now, only if still unused)
       │     └─ already used → roll back → reuse detected (below)
       └─ Insert the new token (parent_id = the used token's ID)
```

Both writes are in one transaction. Otherwise, a failure after marking the
token used but before inserting the new one would burn the client's only token
and force it to log in again.

**Concurrent refreshes with one token.** Step 4 checks `used_at` on a token
read before the transaction, so two requests with the same token can both
pass it. The write settles it: `MarkUsed` updates the row only while `used_at`
is still `NULL`, so exactly one request rotates. The other finds the token
spent, and goes through reuse detection like any replay. See
[decision 0046](../../development/decisions/0046-refresh-token-use-guarded-at-write.md)
and [decision 0047](../../development/decisions/0047-lost-refresh-race-is-reuse.md).

Each new token gets a full `REFRESH_TOKEN_TTL`, so a session in use keeps
moving its expiry forward, and has no maximum age.

## Reuse detection

Reached when `Use` finds the token already spent (someone is presenting a
token that was already exchanged), or when `MarkUsed` finds that a concurrent
request spent it first.

```
  ├─ session.Revoke(now)      (the session loaded in step 2)
  ├─ One transaction: update the session
  └─ return ErrTokenAlreadyUsed → 401
```

**Why revoke the whole session.** A refresh token is used once, so a second
presentation means two parties hold it: the legitimate client, and someone who
copied it. The service cannot tell which one is calling. Revoking the session
locks both out: the attacker for good, the user until they log in again.
Serving the request instead would let an attacker who replayed a token keep a
rotating credential indefinitely. Locking out a user can be recovered from; a
silent, lasting compromise cannot.

The cost is that the legitimate client is logged out too, including when it
is the one sending the token twice, after a lost response or from two tabs.
See [Limitations](../../limitations.md#a-repeated-refresh-logs-the-user-out).

The `parent_id` chain keeps the full rotation history of a compromised
session.

**What revocation does not do.** Access tokens already issued for the session
stay valid until their `exp`, so an attacker locked out of refreshing keeps
access for up to `ACCESS_TOKEN_TTL`. That is the price of verifying tokens
locally. A refresh already running when the session is revoked can also still
finish, with the same effect as if it had finished just before.

## Client obligations

```
1. Store the refresh token privately and durably.
2. After a successful refresh, replace it with the new one at once.
3. Never retry a refresh with the old token: if the response was lost, the
   retry trips reuse detection and logs the user out.
4. Never run two refreshes with the same token at once: across tabs, or
   instances sharing storage, only one may be in flight. The second one trips
   reuse detection too.
5. On a 401 from /v1/auth/refresh, discard the token and send the user to log in.
```

Point 3 is the sharp edge. A client that refreshes, loses the response to a
network error and retries with the same token gets its session revoked. Make
refresh calls one at a time, serialise them across tabs or threads, and log in
again rather than retry blindly.
