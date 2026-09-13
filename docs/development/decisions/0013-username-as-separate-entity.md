# Username as a separate entity

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** domain
- **Related:** [0004](0004-user-and-password-separate-aggregates.md), [0010](0010-username-rules.md)

## Context

`Username` is a value object on `User`. Logging in with email *or* username
would need a user to be reachable through more than one identifier.

The change was considered and deferred: the reshaping is significant, and it
is not clear it pays off until email login actually exists.

## Alternatives considered

- **`Username` as its own entity**, with its own ID and a reference to a user,
  mirroring [`Password`](0004-user-and-password-separate-aggregates.md) —
  would make "log in with email *or* username" natural.
- **Keep `Username` a value object on `User`** until email login exists — the
  refactor is no harder after 1.0 than before.
