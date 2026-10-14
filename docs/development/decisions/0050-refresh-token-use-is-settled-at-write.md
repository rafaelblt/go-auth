# Refresh token use is settled at write, and losing that race is reuse

- **Status:** Accepted

## Invariant

`RefreshTokenRepo.MarkUsed` is the only write to `refresh_tokens.used_at`, it
applies only while the column is `NULL`, and it decides on `RowsAffected`.
`session.ErrTokenAlreadyUsed` revokes the session whatever its origin: no path
answers `401` for a spent token without revoking.

## Context

A refresh token is single use. Refresh reads the token before its unit of work
opens, and `RefreshToken.Use` checks `used_at` in memory, so two requests with
the same token can both read it unused and both pass `Use`. The write used to
be `WHERE id = @id`, so both succeeded, and only the `UNIQUE` on
`refresh_tokens.parent_id` stopped two successors: the second insert failed,
that request answered `500`, and reuse detection never ran.

Two things had to be decided: how the use is persisted, and what the request
that loses is told.

## Decision

`MarkUsed` persists a use already decided in memory, as a
[guarded write](../../architecture/persistence.md#guarded-writes):

```sql
UPDATE refresh_tokens SET used_at = @used_at, updated_at = @updated_at
WHERE id = @id AND used_at IS NULL
```

Under `READ COMMITTED`, a concurrent second update waits for the first to
commit, checks the condition again and matches nothing, so exactly one request
rotates. `Use` stays: it decides expiry, and answers the sequential case
without a write.

**Losing the race is reuse.** When `MarkUsed` returns
`session.ErrTokenAlreadyUsed`, the rotation rolls back, and refresh revokes the
session and answers as for a token found spent on read. Two requests presenting
one token at the same instant are what reuse detection describes: two parties
hold it, and the service cannot tell which is which. Arriving together does not
make either more trustworthy, and if the attacker wins, revoking is the only
thing that locks them out.

## Alternatives

- **The `UNIQUE` on `parent_id` as the guard**: the loser gets a `500` and the
  session is not revoked. The constraint stays, as a structural guarantee that
  the chain cannot fork.
- **A plain `401` for the loser, without revoking**: it would spare a client
  that refreshes from two tabs at once, but if the legitimate client loses and
  does not retry with the same token, an attacker who won keeps a rotating
  credential.
- **Optimistic locking with a `version` column**: generic, and it would also
  serve `Session`, but it needs a migration and a persistence detail carried by
  the entities. Worth revisiting if more entities need concurrency control.

## Consequences

- Of any number of concurrent uses of one token, exactly one rotates, and the
  others revoke the session.
- A client that refreshes concurrently, or retries after a lost response, logs
  its user out everywhere
  ([client obligations](../../api/reference.md#client-obligations)).
