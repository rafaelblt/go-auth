# Migrations

SQL files in `migrations/`, applied with
[golang-migrate](https://github.com/golang-migrate/migrate) and embedded in the
binary with `go:embed`, so a built binary carries its own schema and needs no
files on disk.

```
000001_create_users.up.sql          000001_create_users.down.sql
000002_create_passwords.up.sql      000002_create_passwords.down.sql
000003_create_sessions.up.sql       000003_create_sessions.down.sql
000004_create_refresh_tokens.up.sql 000004_create_refresh_tokens.down.sql
```

Every migration has a `down`. golang-migrate records the applied version in a
`schema_migrations` table it manages itself.

## Version check

`migrations.Latest()` reads the embedded `.up.sql` files and returns the
highest version: the schema this binary was built for.
`postgres.SchemaVersion()` reads the version actually applied, and
`bootstrap.verifySchema` stops startup if the two differ, in either direction
(see [Migrations](../../getting-started.md#migrations)).

`SchemaVersion` tells three failures apart:

| Error | Meaning |
|---|---|
| `ErrSchemaNotInitialized` | No `schema_migrations` table, no row, or a negative version |
| `ErrSchemaDirty` | golang-migrate's `dirty` flag is set: a migration failed halfway, and has to be fixed by hand |
| query error | The database cannot be reached, or the query failed |

A missing table is recognised by PostgreSQL's error code, `42P01`, not by its
message, so a server in another locale does not break the check.

Where this runs during startup: [Startup](../startup.md#composition-root).

## Adding a migration

1. Create `00000N_description.up.sql` and `00000N_description.down.sql`.
2. Check that the `down` undoes the `up`, and on the *right* table. The pair is
   easy to get wrong by copy and paste, and a wrong `down` shows up only during
   a rollback, which is the worst moment.
3. Nothing else needs updating: `migrations.Latest()` and the test helpers pick
   the new file up.
4. Existing deployments must migrate before the new binary will start.
