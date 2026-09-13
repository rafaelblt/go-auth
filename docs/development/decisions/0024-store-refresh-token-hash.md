# Only the refresh token hash is stored

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication
- **Related:** [0023](0023-random-refresh-tokens.md)

## Context

Refresh tokens are credentials, and the database that stores them can leak.

## Decision

The database holds `SHA-256(token)`, never the token. A database leak yields
digests, not usable credentials.

## Alternatives considered

- **Storing the token itself** — a leak would hand out working credentials.
- **bcrypt** — refresh tokens are 32 bytes of uniform randomness, so there is
  nothing to brute-force and no need for a slow KDF. Bcrypt would add hundreds
  of milliseconds per refresh for no security gain, and its input is capped at
  72 bytes anyway.
