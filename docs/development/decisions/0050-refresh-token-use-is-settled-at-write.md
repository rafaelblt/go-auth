# Refresh token use is settled at write, and losing that race is reuse

- **Status:** Accepted
- **Date:** 2026-09-23
- **Supersedes:** [0046](0046-refresh-token-use-guarded-at-write.md), [0047](0047-lost-refresh-race-is-reuse.md)
- **Related:** [Guarded writes](../../architecture/persistence/repositories.md#guarded-writes), [Rotation](../../architecture/usecases/refresh.md#rotation), [Reuse detection](../../architecture/usecases/refresh.md#reuse-detection)

## Invariant

`MarkUsed` is the only write to `refresh_tokens.used_at`, it applies only while
the column is `NULL`, and it decides on `RowsAffected` rather than on the
driver's error. `session.ErrTokenAlreadyUsed` revokes the session whatever its
origin: no path answers `401` for a spent token without revoking.

## Context

A refresh token is single use. Refresh reads the token before its unit of work
opens, `port.UowDeps` holding only writers, and `RefreshToken.Use` checks
`usedAt` in memory. Two requests carrying the same token can both read it
unused, both pass `Use`, and both try to persist the use. This is the situation
[guarded writes](../../architecture/persistence/repositories.md#guarded-writes)
exists for: only the database sees both requests.

Two things had to be decided, and they were originally decided apart, in 0046
and 0047 on the same day. They are one decision, and this record replaces both.

**How the use is persisted.** `RefreshTokenWriter.Update` wrote the token with
`WHERE id = @id`, so the second write also succeeded. The only thing stopping
two successors was the `UNIQUE` constraint on `refresh_tokens.parent_id`: the
second insert failed, the request answered `500`, and reuse detection never ran.
Nothing in the code said the constraint played that role.

**What the losing request is told.** Once the write is guarded, refresh can
learn that a token was already spent in two places: from `RefreshToken.Use`,
when the token was spent before it was read, and from the write, when a
concurrent request spent it in between. Reuse detection covered the first. The
second needed an answer.

## Decision

`port.RefreshTokenWriter.Update` is replaced by `MarkUsed`, which persists the
use of a token already spent in memory and applies only while the stored token
is unused:

```sql
UPDATE refresh_tokens SET used_at = @used_at, updated_at = @updated_at
WHERE id = @id AND used_at IS NULL
```

Under `READ COMMITTED`, a concurrent second update waits for the first to
commit, re-checks the condition and matches nothing. Of any number of concurrent
uses of one token, exactly one rotates.

**Both ways of learning it are reuse.** When `MarkUsed` returns
`session.ErrTokenAlreadyUsed`, the rotation transaction rolls back, refresh
revokes the session and returns `ErrTokenAlreadyUsed`, exactly as for a token
found spent on read.

Two requests presenting the same token at the same instant are the situation
reuse detection describes: two parties hold the token, and the service cannot
tell which is which. Arriving at the same moment instead of one after the other
does not make either more trustworthy. If the attacker wins the race, revoking
is the only thing that locks them out.

**A spent token is told apart from a missing one.** No row updated means either
the token was already used or it does not exist. The two cannot share an answer:
`session.ErrTokenAlreadyUsed` makes the use case revoke the session and report a
reuse. A missing token reported that way would turn a bug — a token read moments
ago that is now gone — into a security response, logging a reuse that never
happened and hiding the actual failure. So `MarkUsed` checks whether the token
exists: a spent token returns `session.ErrTokenAlreadyUsed`, a missing one
returns an unexpected error.

**The existence check runs only after the update.** The update alone decides the
use; the check only explains why it matched nothing. Running it before would cost
a query on every refresh and still be racy, since the row can change between the
two statements. Running it after costs nothing on the common path, and under
`READ COMMITTED` its fresh snapshot already sees the commit that won the race.

**`RefreshToken.Use` stays.** It still decides expiry and answers the sequential
case without a write; the guard is the atomic check behind it. The rule is
stated twice on purpose, which is the shape guarded writes take here.

**A named method, not a guarded `Update`.** The only change a stored refresh
token ever goes through is being spent. A generic `Update` that failed when
`used_at` is set would surprise the next caller; `MarkUsed` says what it does
and when it fails.

## Alternatives considered

- **Keeping the `UNIQUE` on `parent_id` as the guard** — it prevents a forked
  chain, but the loser gets `500` and the session is not revoked. It stays as a
  structural guarantee, not as the reuse check.
- **Answering `401` without revoking, for the request that lost the race** — it
  would spare a client that refreshes from several tabs at once. But if the
  legitimate client loses and does not retry with the same token, an attacker
  who won keeps a rotating credential, which is what reuse detection exists to
  prevent.
- **`SELECT ... FOR UPDATE` inside the unit of work** — correct, but needs
  readers in `port.UowDeps`, which only holds writers, and a longer transaction.
- **Optimistic locking with a `version` column** — generic, and would also serve
  `Session`, but needs a migration and a persistence detail carried by the
  entities. Worth revisiting if more entities need concurrency control.
- **A port-level sentinel instead of the domain one** — it would add a third
  error meaning the same thing next to `session.ErrTokenAlreadyUsed` and
  `refresh.ErrTokenAlreadyUsed`. Where the sentinel lives, and why, is part of
  [guarded writes](../../architecture/persistence/repositories.md#guarded-writes).

## Consequences

- Of any number of concurrent uses of one token, exactly one rotates. The others
  are answered as reuse, and the session is revoked.
- A client that sends concurrent refreshes with one token is logged out, in
  every tab, and so is one that retries after a lost response. Clients must
  serialize refreshes; see
  [client obligations](../../architecture/usecases/refresh.md#client-obligations).
  Whether the service should tolerate legitimate duplicates is left open, in
  [Limitations](../../limitations.md#a-repeated-refresh-logs-the-user-out).
- The not-found check adds a query, only on the path where no row was updated.
- `SessionWriter.Update` is still unguarded: two concurrent revocations both
  write, and the later one overwrites `revoked_at`. Harmless today, since both
  write the same outcome.
