# Persistence

`internal/infra/postgres`, `internal/infra/migrate`, `migrations`

PostgreSQL, through `pgx/v5` and `pgxpool`. The repositories implement the
`internal/port` interfaces and are the only code that writes SQL.

## Connection pool

`NewPool` rejects an empty or blank connection string itself, because pgx
would accept it and fall back to the libpq environment variables, connecting to
whatever database those name. It sets the session timezone to UTC and decodes
`timestamptz` in UTC, so with `SystemClock` also in UTC, every time is UTC
whatever the server's locale. It pings the database, so a bad connection fails
at startup. Pool sizing is pgx's default, which depends on the CPU count, and is
not configurable.

## Schema

```
users ──1:0..1── passwords
  │
  └──1:N── sessions ──1:N── refresh_tokens ── parent_id ──► refresh_tokens

signing_keys   (no foreign key to or from any other table)
```

The SQL is in [`migrations/`](../../migrations). What the constraints are for:

- **`users.username`** is `UNIQUE`, and that index decides two concurrent
  registrations of one name ([Guarded writes](#guarded-writes)). Values are
  lower case: the domain canonicalises them first.
- **`passwords.user_id`** is `UNIQUE`: one password per user, which the domain
  does not enforce. **`passwords.hash`** holds the whole bcrypt string, with its
  cost and salt, so each password keeps the cost it was hashed with: changing
  `BCRYPT_COST` breaks no login, and nothing rehashes a password until the user
  changes it.
- **`sessions.revoked_at`** is `NULL` while the session is active: a timestamp
  rather than a flag, so the time is recorded.
- **`refresh_tokens.hash`** is the raw SHA-256 digest, never the token, and
  `UNIQUE`, which gives the lookup its index. **`parent_id`** is `UNIQUE`, so a
  token has at most one successor and the rotation chain cannot fork.
  **`used_at`** is `NULL` until the token is spent.
- **`signing_keys.generation`** orders the keys and, as the primary key,
  settles two instances adding the same one. **`seed`** is the Ed25519 private
  key, 32 bytes as it is or 60 bytes sealed
  ([Sealed seeds](tokens.md#sealed-seeds)). **`active_at`** is when the key
  starts signing, and **`created_at`** when it was added and published. The key
  ID is derived from the key, so it is not stored.

Nothing deletes users, sessions or refresh tokens
([Limitations](../limitations.md#session-and-refresh-token-rows-are-never-deleted));
the keyring deletes retired signing keys.

## Migrations

The `.sql` files in `migrations/` are embedded in the binary with `go:embed`
and applied with [golang-migrate](https://github.com/golang-migrate/migrate),
which records the version in its own `schema_migrations` table. Every migration
has a `down`.

At startup, `bootstrap.verifySchema` compares `migrations.Latest()`, the highest
embedded version, with `postgres.SchemaVersion()`, the one applied, and stops
the service if they differ in either direction: a binary running against a
schema it was not built for would fail one request at a time instead.
`SchemaVersion` tells three failures apart:

| Error | Meaning |
|---|---|
| `ErrSchemaNotInitialized` | No `schema_migrations` table, no row, or a negative version |
| `ErrSchemaDirty` | A migration failed halfway, and has to be fixed by hand |
| query error | The database cannot be reached, or the query failed |

A missing table is recognised by its SQLSTATE, `42P01`, not by its message, so
a server in another locale does not break the check. Applying migrations, as an
operator: [Getting started](../getting-started.md#migrations).

To add one, create the `up` and `down` pair with the next number, and check
that the `down` undoes the `up` on the right table: a wrong `down` only shows up
during a rollback. `Latest()` and the test helpers pick it up. A released
migration is never edited; a change is a new migration.

## Repositories

One per table, each taking a `DB` interface (`QueryRow`, `Query`, `Exec`)
that both `*pgxpool.Pool` and `pgx.Tx` satisfy, so the same repository serves a
plain query and a write inside a transaction.

| Repository | Implements |
|---|---|
| `UserRepo` | `UserReader`, `UserWriter`, `UserExistsChecker` |
| `PasswordRepo` | `PasswordReader`, `PasswordWriter` |
| `SessionRepo` | `SessionReader`, `SessionWriter` |
| `RefreshTokenRepo` | `RefreshTokenReader`, `RefreshTokenWriter` |
| `SigningKeyRepo` | `SigningKeyStore` |

`models.go` holds a model per table, with `db:` tags, and the mapping both
ways; entities carry no database tags. A row becomes an entity through the
domain's constructors ([Constructors](conventions.md#constructors)), so a row
that breaks a rule is an error, not a malformed entity. Inserts use
`pgx.StrictStructArgs`, which fails when a named parameter and a model field do
not match, so a column added on one side only fails loudly. Updates name only
the columns they set: `SessionRepo.Update` with `pgx.StructArgs`, `UpdateHash`
with `pgx.NamedArgs`, `RevokeAllByUserID` with positional arguments.

A `Find*` that finds nothing returns `(nil, nil)`: whether a missing row is a
problem is the caller's decision. An update that matches no row is an error,
since the row vanished; `RevokeAllByUserID` is the exception, as a user may
have no session left to revoke.

## Guarded writes

How the code settles every rule that has to hold across concurrent requests.

**The premise.** Under PostgreSQL's default isolation, `READ COMMITTED`, a use
case that reads and then writes cannot make the pair atomic: two requests can
read the same state and both decide to act. Only the database sees both.

**The shape.** The rule is stated twice. A check in memory, on what was read
before the transaction, answers the common case but decides nothing. Then the
statement itself carries the condition, in a `WHERE` or an `ON CONFLICT`, so the
database applies it only while the rule still holds:

| Statement | Condition | When nothing matched |
|---|---|---|
| `UserRepo.Add` | `ON CONFLICT (username) DO NOTHING` | `user.ErrUsernameAlreadyExists`: register's `409` |
| `RefreshTokenRepo.MarkUsed` | `WHERE id = @id AND used_at IS NULL` | `session.ErrTokenAlreadyUsed`: reuse ([Rotation](usecases.md#rotation)) |
| `PasswordRepo.UpdateHash` | `WHERE id = @id AND hash = @previous_hash` | `password.ErrHashChanged`: `401`, `password changed` |
| `PasswordRepo.LockHash` | `WHERE id = @id AND hash = @hash FOR SHARE` | `password.ErrHashChanged`: `401`, `password changed` ([The password lock](usecases.md#the-password-lock)) |
| `SigningKeyRepo.Add` | `ON CONFLICT (generation) DO NOTHING` | nothing: the keyring lists the keys again ([Signing keys](tokens.md#signing-keys)) |

`LockHash` takes a lock instead of writing, for a rule that spans two tables
([The password lock](usecases.md#the-password-lock)). `FOR SHARE` is the
weakest lock that works: the `UPDATE` of a password change takes
`FOR NO KEY UPDATE`, which conflicts with `FOR SHARE` but not with
`FOR KEY SHARE`, and `FOR UPDATE` would also make the logins of one user wait
for each other.

**The outcome comes from `RowsAffected`**, never from the driver's error.
Matching SQLSTATE `23505` would also need the constraint's name, which
PostgreSQL generated (`users_username_key`) and nothing in the project
declares, and would absorb conflicts the repository did not mean to. Naming the
column absorbs exactly one: a primary key collision stays an error.

**A statement that matched nothing checks why.** No row changed means either
the rule no longer holds or the row does not exist, and only the first is the
domain error. `MarkUsed`, `UpdateHash` and `LockHash` then look the row up by
ID, and report a missing one as an unexpected error, so a bug is not answered as
a lost race. The lookup runs only on that path, and under `READ COMMITTED` it
already sees the commit that won.

**The sentinel is the domain's**, not the port's: the repository reports a fact
about stored state, and what to do about it stays in the use case. The port
documents it on the method. The methods say what they do, `MarkUsed` and an
`UpdateHash` that takes the hash it replaces, rather than a generic `Update`
that would refuse silently.

Why not the alternatives, for all of them: reading inside the transaction with
`SELECT … FOR UPDATE` needs readers in the unit of work and holds a lock across
whatever comes between, bcrypt included; `SERIALIZABLE` needs every
serialization failure retried; an advisory lock is a second pattern held across
round trips. A new rule of this kind takes the same shape.

## Unit of work

Use cases reach transactions through `port.UnitOfWork`:

```go
err := uow.Do(ctx, func(deps port.UowDeps) error {
    if err := deps.SessionWriter.Add(ctx, sess); err != nil {
        return err
    }
    return deps.RefreshTokenWriter.Add(ctx, token)
})
```

`Do` begins a transaction, builds repositories bound to it, runs the callback
and commits; any error rolls everything back. `UowDeps` holds only writers, so
reads happen before the transaction opens, which keeps transactions short, and
a write that depends on what was read checks again itself, as a guarded write.
`LockHash` reads inside the transaction only to lock.
