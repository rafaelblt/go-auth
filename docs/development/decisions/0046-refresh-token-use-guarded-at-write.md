# Refresh token use is guarded at write

- **Status:** Accepted
- **Date:** 2026-09-13
- **Areas:** authentication, architecture
- **Related:** [0047](0047-lost-refresh-race-is-reuse.md), [Rotation](../../architecture/usecases/refresh.md#rotation), [Reuse detection](../../architecture/usecases/refresh.md#reuse-detection)

## Context

A refresh token is single-use. Refresh reads the token before its unit of work
opens (`port.UowDeps` holds only writers), and `RefreshToken.Use` checks
`usedAt` in memory. Two requests with the same token can both read it unused,
both pass `Use`, and both try to persist the use.

Before this record, `RefreshTokenWriter.Update` wrote the token with
`WHERE id = @id`, so the second write also succeeded. The only thing stopping
two successors was the `UNIQUE` constraint on `refresh_tokens.parent_id`: the
second insert failed, the request answered `500`, and reuse detection never
ran. Nothing in the code said the constraint played that role.

In memory, the check cannot be made atomic across requests. Only the database
sees both.

## Decision

`port.RefreshTokenWriter.Update` is replaced by `MarkUsed`. It persists the
use of a token already spent in memory, and applies only while the stored
token is unused:

```sql
UPDATE refresh_tokens SET used_at = @used_at, updated_at = @updated_at
WHERE id = @id AND used_at IS NULL
```

Under `READ COMMITTED`, a concurrent second update waits for the first to
commit, re-checks the condition and matches nothing.

**A spent token is told apart from a missing one.** No row updated means
either the token was already used or it does not exist. The two cannot share
an answer: `session.ErrTokenAlreadyUsed` makes the use case revoke the session
and report a reuse. A missing token reported that way would turn a bug — a
token read moments ago that is now gone — into a security response, logging a
reuse that never happened and hiding the actual failure. So `MarkUsed` checks
whether the token exists: a spent token returns `session.ErrTokenAlreadyUsed`,
a missing one returns an unexpected error.

**The existence check runs only after the update.** The update alone decides
the use; the check only explains why it matched nothing. Running it before
would cost a query on every refresh and still be racy, since the row can
change between the two statements. Running it after costs nothing on the
common path, and under `READ COMMITTED` its fresh snapshot already sees the
commit that won the race.

`RefreshToken.Use` stays. It still decides expiry and answers the sequential
case without a write; the guard is the atomic check behind it. The rule is
stated twice on purpose.

**A named method, not a guarded `Update`.** The only change a stored refresh
token ever goes through is being spent. A generic `Update` that fails when
`used_at` is set would surprise the next caller; `MarkUsed` says what it does
and when it fails.

**The repository returns the domain error.** "This token was already used" is
the same fact whether `Use` finds it in memory or the database finds it at
write time, so the use case handles both with one branch. The repository
reports a fact about stored state; what to do about it stays in the use case
(see [0047](0047-lost-refresh-race-is-reuse.md)). A port-level sentinel would
add a third error with the same meaning next to `session.ErrTokenAlreadyUsed`
and `refresh.ErrTokenAlreadyUsed`.

## Alternatives considered

- **Keeping the `UNIQUE` on `parent_id` as the guard** — it prevents a forked
  chain, but the loser gets `500` and the session is not revoked. It stays as
  a structural guarantee, not as the reuse check.
- **`SELECT ... FOR UPDATE` inside the unit of work** — correct, but needs
  readers in `port.UowDeps`, which only holds writers, and a longer
  transaction.
- **Optimistic locking with a `version` column** — generic, and would also
  serve `Session`, but needs a migration and a persistence detail carried by
  the entities. Worth revisiting if more entities need concurrency control.
- **A sentinel in `internal/port`** — see above.

## Consequences

- Of any number of concurrent uses of one token, exactly one rotates.
- The not-found check adds a query, only on the path where no row was
  updated.
- `SessionWriter.Update` is still unguarded: two concurrent revocations both
  write, and the later one overwrites `revoked_at`. Harmless today.
