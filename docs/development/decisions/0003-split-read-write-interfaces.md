# Split read and write interfaces

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** architecture
- **Related:** [0002](0002-ports-and-adapters.md)

## Context

A repository such as `postgres.UserRepo` reads users, writes them and checks
whether one exists. Each use case needs only some of that.

## Decision

`UserReader`, `UserWriter` and `UserExistsChecker` are separate interfaces even
though `postgres.UserRepo` implements all three.

Each use case then declares only what it uses, and its `Config` struct becomes
an honest statement of its capabilities.

## Alternatives considered

- **One interface per repository** — every use case would receive every
  capability of the repository, whether it uses it or not.

## Consequences

`register.Config` has a `UserExistsChecker` and no `UserReader` — registration
cannot read users, and that is visible without reading the implementation.
