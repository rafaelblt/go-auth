# Database configuration in parts

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

The database is configured only through `DATABASE_URL`. The proposal is to
accept `PG_USER`, `PG_PASSWORD`, `PG_HOST` and so on alongside it. The open
question is how `Config` represents the result.

## Alternatives considered

- **Keep only the URL in `Config` and assemble it** — makes `config`
  responsible for URL construction.
- **Hold both shapes** and let a later layer decide.
- **A dedicated database configuration type** that is one or the other — the
  most idiomatic, and the least obvious to implement.

Whichever representation is chosen, providing both forms should be an error
rather than a precedence rule — an operator should never have to guess which
one won.
