# A duplicate username is reported by the user writer

- **Status:** Accepted
- **Date:** 2026-09-17
- **Related:** [0046](0046-refresh-token-use-guarded-at-write.md), [Register](../../architecture/usecases/register.md)

## Invariant

`UserRepo.Add` applies only while the username is free, through
`ON CONFLICT (username) DO NOTHING`, and returns
`user.ErrUsernameAlreadyExists` when it inserted nothing. It decides on
`RowsAffected`, never on the driver's error, and it depends on the unique index
on `users.username`.

## Context

Register checks whether the username is free, hashes the password, then
inserts the user and its password in one transaction. Under PostgreSQL's
default isolation, `READ COMMITTED`, two registrations of the same username
arriving together both pass the check: neither transaction sees the other's
row before it commits. Only the `UNIQUE` constraint on `users.username` sees
both.

Before this record nothing translated that constraint violation. The insert
failed, register treated the error as unexpected, and the loser of the race
got `500 internal_server_error` instead of `409 username_already_exists`. No
duplicate user was created, so the outcome was safe and the answer was wrong.

Register used to recognise a "username already exists" error from the writer,
and lost that check when it moved to its own package (f9a694d). No repository
ever returned such an error, so the check only ever worked against fakes.

In memory, the check cannot be made atomic across requests. Only the database
sees both.

## Decision

`port.UserWriter.Add` applies only while the username is free, and returns
`user.ErrUsernameAlreadyExists` when another insert took it first. Register
translates that into `register.ErrUsernameAlreadyExists`, the same `409` the
early check produces.

**The insert takes the conflict itself.**

```sql
INSERT INTO users (id, username, status, created_at, updated_at)
VALUES (@id, @username, @status, @created_at, @updated_at)
ON CONFLICT (username) DO NOTHING
```

No row inserted means the username is taken. The repository decides on
`RowsAffected`, not on the driver's error, which matters for three reasons:

- **The constraint has no name anyone chose.** The column is declared
  `username TEXT UNIQUE NOT NULL`, so PostgreSQL generates
  `users_username_key`. Matching SQLSTATE `23505` would also have to match
  that name, because the primary key is unique too, and would tie the
  repository to a name nothing in the project declares. `ON CONFLICT` names
  the column instead.
- **It absorbs one conflict, not every conflict.** Targeting `(username)`
  leaves a primary key collision as an error, which is what a repeated `ID`
  should be.
- **It is the shape the repositories already use.** A guarded write decided by
  `RowsAffected` is what `RefreshTokenWriter.MarkUsed` does
  ([0046](0046-refresh-token-use-guarded-at-write.md)).

Nothing inspects `pgconn.PgError` outside the startup schema check, so there
is still only one such place and no helper to extract for it.

**The sentinel lives in `internal/domain/user`.** It is the same placement
[0046](0046-refresh-token-use-guarded-at-write.md) chose for
`session.ErrTokenAlreadyUsed`: the repository reports a fact about stored
state, and what to do about it stays in the use case. Putting this one in
`internal/port` instead would leave two rules of the same shape in two
different packages, and `port` holds the abstractions, not the vocabulary.

The placement is not symmetrical with `session.ErrTokenAlreadyUsed`, and the
difference is worth stating: that one is also produced by `RefreshToken.Use`,
whereas nothing in `internal/domain/user` produces or consumes this error —
a single `User` cannot enforce uniqueness across users. It is still a fact
about the domain rather than about one writer's contract, and any other
implementation of `UserWriter` reports the same fact. The port documents it
on `Add`, exactly as it documents `session.ErrTokenAlreadyUsed` on `MarkUsed`.

**The early existence check stays.** Not to save building two entities, which
costs nothing, but to save a bcrypt hash: without it, registering an existing
username repeatedly makes an unauthenticated caller spend the service's CPU on
a full hash each time. The check replaces that with one lookup on a unique
index. As in [0046](0046-refresh-token-use-guarded-at-write.md), the rule is
stated twice on purpose: the check answers the common path cheaply, the insert
is the atomic decision behind it.

## Alternatives considered

- **A sentinel in `internal/port`** — see above.
- **Inspecting `pgconn.PgError` for `23505` plus the constraint name** — see
  above.
- **Dropping the early check and relying only on the insert** — one query
  fewer and a single path, but it hashes a password for every attempt on a
  taken username.
- **Moving the check into the transaction** — no help under `READ COMMITTED`:
  neither transaction sees the other's row.
- **`SERIALIZABLE`, or an advisory lock on the username** — correct, and far
  more than a rare conflict on one column justifies.
- **Mapping the `500` to a `409` in `internal/api`** — it would put a
  PostgreSQL detail in the HTTP layer and answer `409` for unrelated failures.

## Consequences

- Of any number of concurrent registrations of one username, exactly one
  creates a user; the others get `409`.
- Register makes two queries on the common path, the existence check and the
  insert.
- `ON CONFLICT (username)` requires the unique index on `users.username`.
  Dropping it does not make duplicates pass silently: PostgreSQL rejects the
  statement outright, so the tests fail loudly.
- The password is never inserted for a username that was taken in the
  meantime: `Add` returns first and the unit of work rolls the transaction
  back.
