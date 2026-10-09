# Session and refresh tokens

`internal/domain/session`

One package, because a session and its refresh tokens change together:
revoking a session invalidates its tokens, and reuse detection reads a token
and revokes its session in one operation.

## Session

`internal/domain/session/session.go`

A login. Created by `/v1/auth/login`, revoked by reuse detection.

| Field                     | Type         | Notes              |
| ------------------------- | ------------ | ------------------ |
| `id`                      | `SessionID`  |                    |
| `userID`                  | `user.ID`    |                    |
| `revokedAt`               | `*time.Time` | `nil` while active |
| `createdAt` / `updatedAt` | `time.Time`  |                    |

`Revoke(at)` is idempotent in memory: revoking a session that is already
revoked changes nothing, so a second call cannot overwrite the time of the
first. The guarantee stops at the entity. `SessionWriter.Update` is not
guarded, so two concurrent requests that each read an active session and
revoke it both write, and the later write wins; see
[decision 0050](../../development/decisions/0050-refresh-token-use-is-settled-at-write.md#consequences).

`SessionWriter.RevokeAllByUserID` revokes every active session of a user in
one statement, without reading them into entities, so it reaches sessions the
caller never loaded. It sets `revoked_at` only where it is still `NULL`, which
is the rule of `Revoke` stated in SQL: a session revoked earlier keeps its
time.

## RefreshToken

`internal/domain/session/refresh_token.go`

One link in a rotation chain.

| Field                     | Type               | Notes                                              |
| ------------------------- | ------------------ | -------------------------------------------------- |
| `id`                      | `RefreshTokenID`   |                                                    |
| `sessionID`               | `SessionID`        | the session it belongs to                          |
| `hash`                    | `RefreshTokenHash` | SHA-256 of the secret                              |
| `parentID`                | `*RefreshTokenID`  | the token this one replaced; `nil` for the first   |
| `expiresAt`               | `time.Time`        |                                                    |
| `usedAt`                  | `*time.Time`       | `nil` until spent                                  |
| `createdAt` / `updatedAt` | `time.Time`        |                                                    |

`NewRefreshToken` generates the token's secret and returns it with the token.
Its params have no hash field, so a new token's hash always comes from a real
secret. The token keeps only the hash, so the returned secret is the one
chance to hand it to the client. Why the secret belongs to the domain:
[decision 0049](../../development/decisions/0049-refresh-token-secret-in-the-domain.md).

`parentID` makes the chain explicit: login creates a token without a parent,
and each refresh creates one pointing at the token it replaced. The chain is
the whole history of a session's rotations.

### `Use`

Spends the token, and is the heart of reuse detection:

```go
func (t *RefreshToken) Use(usedAt time.Time) error {
    if t.usedAt != nil {
        return ErrTokenAlreadyUsed
    }
    if usedAt.After(t.expiresAt) {
        return ErrTokenExpired
    }
    t.usedAt = &usedAt
    t.updatedAt = usedAt
    return nil
}
```

The order matters. "Already used" is checked before "expired", so replaying a
token that was spent *and* has since expired still counts as reuse, and still
revokes the session. Checked the other way round, someone holding an old
stolen token could wait for it to expire, replay it, get an ordinary "expired",
and never trip the alarm.

`Use` decides in memory, on a token read before the write. Two concurrent
refreshes can both pass it, so the write checks again; see
[decision 0050](../../development/decisions/0050-refresh-token-use-is-settled-at-write.md)
and [Refresh](../usecases/refresh.md).

## RefreshTokenSecret

`internal/domain/session/refresh_token_secret.go`

The credential the client holds: 32 bytes from `crypto/rand`, which
`Value()` encodes in base64url. Only `NewRefreshToken` creates one.
`ParseRefreshTokenSecret` reads a client's string back, rejecting a bad
encoding or length, and `Hash()` returns the `RefreshTokenHash` to look the
token up by. It exposes `Value()`, not `String()`.

Both methods panic on a zero secret. One can only come from an ignored error,
and its hash would otherwise be looked up as a token that simply does not
exist.

## RefreshTokenHash

`internal/domain/session/refresh_token_hash.go`

The SHA-256 digest of a secret, exactly 32 bytes. Both the constructor and
`Value()` copy the bytes. The database stores only this hash, never the
secret.

## AccessToken

`internal/domain/session/access_token.go`

A signed JWT string, not empty. It is a value object rather than a `string` so
it cannot be mixed up with another string on its way out. It exposes
`Value()`, not `String()`.

What the two token types are for, and how they are signed:
[Tokens](../tokens.md).
