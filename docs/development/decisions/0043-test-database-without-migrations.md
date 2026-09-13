# The test database applies no migrations

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** testing

## Context

Migration tests need a database their migrations have not already been
applied to.

## Decision

`testutil.NewDatabase` returns an empty database. Migrations are applied by the
separate `migratetest` helper when a test wants them.

## Alternatives considered

- **Applying migrations in `NewDatabase`** — the original behaviour. It made the
  migration tests themselves impossible to write.
