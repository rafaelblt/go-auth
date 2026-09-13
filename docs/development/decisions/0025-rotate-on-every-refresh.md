# Rotation on every refresh

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication
- **Related:** [0026](0026-reuse-revokes-session.md)

## Context

A refresh token that stays valid until it expires is valuable to whoever holds
a copy for that whole time, and a replayed copy is indistinguishable from the
original.

## Decision

Every refresh issues a new refresh token and marks the presented one used, with
`parent_id` linking the chain.

## Consequences

Rotation bounds how long any single token is valuable and makes reuse
detectable at all — the property the whole scheme rests on. See
[Reuse detection revokes the whole session](0026-reuse-revokes-session.md).
