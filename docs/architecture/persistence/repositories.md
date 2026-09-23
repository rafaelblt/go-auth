# Repositories

`internal/infra/postgres`

One per entity, each taking a `DB` interface rather than a pool:

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
an error: updating a row that vanished is a bug, not a no-op.

**Guarded writes are decided by the rows they change.** `UserRepo.Add` and
`RefreshTokenRepo.MarkUsed` apply only while a condition holds, and report
the domain error (`user.ErrUsernameAlreadyExists`,
`session.ErrTokenAlreadyUsed`) when nothing changed, instead of inspecting the
driver's error.

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
