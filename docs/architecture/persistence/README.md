# Persistence

PostgreSQL, reached through `pgx/v5` and `pgxpool`. The repositories
implement the `internal/port` interfaces, and are the only code that writes
SQL.

| Document | Covers |
|---|---|
| [Schema](schema.md) | The five tables, and what each constraint is for |
| [Migrations](migrations.md) | The embedded `.sql` files, the version check, adding one |
| [Repositories](repositories.md) | Repositories, mapping, conventions, guarded writes, the unit of work |

## Connection pool

`internal/infra/postgres/pool.go`

`NewPool` parses the URL, sets the session timezone to UTC, registers a
`timestamptz` codec that decodes in UTC, and pings the database, so a bad
connection fails at startup.

It rejects an empty connection string itself, because pgx would accept it and
fall back to the libpq environment variables, connecting to whatever database
those name. Pool sizing uses the pgx defaults; see
[`DATABASE_URL`](../../configuration.md#database_url).

## Time and timezones

Everything is UTC, enforced in three places: `SystemClock.Now()` returns
`time.Now().UTC()`, the pool sets the session timezone to UTC, and the
`timestamptz` codec decodes into UTC. Every timestamp column is
`TIMESTAMPTZ`.
