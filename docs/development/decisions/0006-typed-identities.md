# Typed identities rather than bare UUIDs

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain

## Context

Four entities all have UUID identities, and functions take several at once.

## Decision

`user.ID`, `session.SessionID` and `password.ID` all wrap `shared.EntityID`,
which wraps a UUID. Passing a `user.ID` where a `SessionID` is expected is a
compile error.

## Alternatives considered

- **Bare `uuid.UUID` or `string` identifiers** — the compiler cannot help.
  Swapping two identifiers is then an easy mistake, and it produces a runtime
  "not found" rather than a build failure.
