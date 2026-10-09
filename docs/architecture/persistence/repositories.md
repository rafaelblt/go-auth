# Repositories

`internal/infra/postgres`

One per table, each taking a `DB` interface rather than a pool:

```go
type DB interface {
    QueryRow(ctx, sql string, args ...any) pgx.Row
    Query(ctx, sql string, args ...any) (pgx.Rows, error)
    Exec(ctx, sql string, args ...any) (pgconn.CommandTag, error)
}
```

Both `*pgxpool.Pool` and `pgx.Tx` satisfy it, so the same repository type
serves a plain query and a write inside a transaction.

| Repository | Implements |
|---|---|
| `UserRepo` | `UserReader`, `UserWriter`, `UserExistsChecker` |
| `PasswordRepo` | `PasswordReader`, `PasswordWriter` |
| `SessionRepo` | `SessionReader`, `SessionWriter` |
| `RefreshTokenRepo` | `RefreshTokenReader`, `RefreshTokenWriter` |
| `SigningKeyRepo` | `SigningKeyStore` |

The interfaces themselves are declared in
[`internal/port`](../usecases/README.md#ports).

## Models and mapping

`models.go` holds a model struct per table, with `db:` tags, and the mapping
functions in both directions. Domain entities never see SQL and carry no
database tags.

An entity becomes a model through its accessors. A model becomes an entity
through the `Restore*` constructors, so a row that breaks an invariant
produces a clear error instead of a malformed entity.

Inserts use `pgx.StrictStructArgs`: every named parameter in the SQL must be
in the struct, so a column added to the model but not to the statement fails
loudly. Updates use `pgx.StructArgs`, which allows struct fields the
statement does not name, since an update touches only some columns.

## Conventions

**"Not found" is `(nil, nil)`, not an error.** Every `Find*` returns a `nil`
entity and a `nil` error when the row does not exist, so callers write
`if usr == nil` instead of matching a sentinel. Whether a missing row is a
problem is the caller's decision.

**Updates check that they changed a row.** An update that matches no row is
an error: updating a row that vanished is a bug, not a no-op. The exception is
`SessionRepo.RevokeAllByUserID`, which updates every active session of a user
in one statement, so it matches as many rows as the user has active sessions,
none included. It sets `revoked_at` only where it is `NULL`, so a session
revoked earlier keeps its time, as `Session.Revoke` does in memory.

**Guarded writes are decided by the rows they change.** `UserRepo.Add`,
`RefreshTokenRepo.MarkUsed` and `PasswordRepo.UpdateHash` apply only while a
condition holds, and report the domain error
(`user.ErrUsernameAlreadyExists`, `session.ErrTokenAlreadyUsed`,
`password.ErrHashChanged`) when nothing changed, instead of inspecting the
driver's error. The pattern has its own section, below.
`SigningKeyRepo.Add` applies only while the generation is free
(`ON CONFLICT (generation) DO NOTHING`) and reports nothing when it is taken,
because its caller reads the keys again either way.

## Guarded writes

The pattern above is worth stating on its own, because it is how this codebase
answers every rule that has to hold across concurrent requests, and because the
next such rule should reuse it rather than invent a shape.

**The premise.** Under PostgreSQL's default isolation, `READ COMMITTED`, a use
case that reads and then writes cannot make the pair atomic: two requests can
both read the same state and both decide to act. In memory the check cannot be
made atomic across requests. Only the database sees both.

**The shape.** The rule is stated twice, on purpose:

1. **A cheap check in memory**, before the write. It answers the common case
   without a transaction — `register` asks `ExistsByUsername` before spending a
   bcrypt hash, `refresh` calls `RefreshToken.Use` before opening its unit of
   work. This check decides nothing; it saves work on the path that is almost
   always taken.
2. **A guarded write**, which is the actual decision. The `WHERE` clause or the
   `ON CONFLICT` carries the condition, so the database applies the statement
   only while the rule still holds:

   ```sql
   UPDATE refresh_tokens SET used_at = @used_at, updated_at = @updated_at
   WHERE id = @id AND used_at IS NULL

   INSERT INTO users (...) VALUES (...) ON CONFLICT (username) DO NOTHING

   UPDATE passwords SET hash = @hash, updated_at = @updated_at
   WHERE id = @id AND hash = @previous_hash
   ```

**The outcome is read from `RowsAffected`, never from the driver's error.** No row
changed means the rule no longer holds. Matching a SQLSTATE plus a constraint
name would tie the repository to a name nothing in the project declares —
PostgreSQL generates `users_username_key` from the column — and would catch
conflicts the repository did not mean to absorb. Targeting the column instead
absorbs exactly one conflict and leaves a primary key collision an error.

**The sentinel lives in the domain package**, not in `internal/port`:
`user.ErrUsernameAlreadyExists`, `session.ErrTokenAlreadyUsed`,
`password.ErrHashChanged`. The repository
reports a fact about stored state; what to do about that fact stays in the use
case. `port` holds the abstractions, not the vocabulary, and it documents the
error on the method that can return it.

**One guarded write reports nothing.** `SigningKeyRepo.Add` targets its column
like the others, with `ON CONFLICT (generation) DO NOTHING`, but reads no
`RowsAffected` and returns no sentinel. Its only caller, the keyring, lists the
keys again after every `Add`, because it needs the key that is stored, not who
stored it
([decision 0053](../../development/decisions/0053-signing-keys-are-shared-through-the-database.md)).

**A named method, not a guarded `Update`.** `MarkUsed` says what it does and when
it fails. A generic `Update` that silently refused when `used_at` was set would
surprise its next caller. `UpdateHash` takes the hash it replaces as an
argument for the same reason: the condition is in the signature, not hidden in
the statement.

**A guarded update that matches nothing checks why.** No row changed means
either the rule no longer holds or the row does not exist, and only the first
is the domain error. `MarkUsed` and `UpdateHash` look the row up by ID after
the update, and report a missing one as an unexpected error, so a bug is not
answered as a lost race. The lookup runs only on that path, and under
`READ COMMITTED` it already sees the commit that won.

Where each of these was decided, with the alternatives weighed:
[decision 0050](../../development/decisions/0050-refresh-token-use-is-settled-at-write.md)
and [decision 0048](../../development/decisions/0048-duplicate-username-reported-by-the-writer.md).
A new rule of this shape should link here rather than argue the premise again.

## Unit of work

`internal/infra/postgres/uow.go`

Use cases reach transactions through `port.UnitOfWork`, which knows nothing
about SQL:

```go
err := uow.Do(ctx, func(deps port.UowDeps) error {
    if err := deps.SessionWriter.Add(ctx, sess); err != nil {
        return err
    }
    return deps.RefreshTokenWriter.Add(ctx, token)
})
```

`Do` begins a transaction, builds a fresh set of repositories **bound to
it**, runs the callback, and commits. Any error from the callback rolls the
whole transaction back. The rollback is deferred on every path; after a
commit, pgx makes it a no-op.

`UowDeps` holds only writers, so a use case cannot mix in a repository
outside the transaction, and cannot read inside it either. Reads happen
before the transaction opens, which keeps transactions short. When a write
depends on what was read, the write checks again itself, as the guarded writes
above do: under `READ COMMITTED`, a read inside the transaction would not see
a concurrent transaction's uncommitted row anyway.
