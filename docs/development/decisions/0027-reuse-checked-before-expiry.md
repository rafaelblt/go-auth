# Reuse is checked before expiry

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication
- **Related:** [0026](0026-reuse-revokes-session.md)

## Context

A replayed refresh token can be both spent and expired, and each condition
leads to a different response.

## Decision

`RefreshToken.Use` checks `usedAt` before `expiresAt`. Replaying a token that is
both spent and expired still triggers session revocation.

## Alternatives considered

- **Checking expiry first** — an attacker with an old stolen token could wait
  for it to expire, replay it, and get an ordinary "expired" response without
  tripping the alarm, losing the signal that a token was compromised at all.
