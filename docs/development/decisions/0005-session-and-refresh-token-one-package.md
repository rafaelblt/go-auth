# `Session` and `RefreshToken` in one package

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain, authentication
- **Related:** [0026](0026-reuse-revokes-session.md)

## Context

Reuse detection reads a token and revokes a session in the same operation;
revoking a session invalidates its tokens.

## Decision

`Session` and `RefreshToken` live together in `internal/session`. They are one
consistency boundary.

## Alternatives considered

- **Separate packages** — would put a circular dependency between two packages
  that always change together.
