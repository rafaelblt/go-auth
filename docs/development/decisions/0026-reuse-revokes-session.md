# Reuse detection revokes the whole session

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication
- **Related:** [0005](0005-session-and-refresh-token-one-package.md), [0025](0025-rotate-on-every-refresh.md), [0027](0027-reuse-checked-before-expiry.md)

## Context

A single-use token presented twice means two parties hold it, and the service
cannot tell which one is calling.

## Decision

Presenting an already-used refresh token revokes its whole session. That locks
out both parties: the attacker permanently, the legitimate user until they log
in again.

## Alternatives considered

- **Serving the request** — would leave an attacker who replayed a token
  holding a rotating credential indefinitely.

## Consequences

A legitimate user is logged out whenever a token from their session is
replayed. Locking out a legitimate user is recoverable; silent, persistent
compromise is not.
