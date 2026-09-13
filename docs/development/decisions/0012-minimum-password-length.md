# Minimum password length

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** domain
- **Related:** [0011](0011-password-input-rules.md)

## Context

The minimum is currently 4 code points, set by
[Password input rules](0011-password-input-rules.md). That is a floor on "is
this a password at all", not a security recommendation.

Worth settling before 1.0, since raising it later is a policy change existing
users would feel at their next password change.

## Alternatives considered

- **Keep 4 code points** — the current floor.
- **Raise to 8 code points** — the defensible security minimum.
