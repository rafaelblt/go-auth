# Minimal JWT claims

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication

## Context

The service serves one application per deployment.

## Decision

Access tokens carry only `sub` and `exp`. No `iss`, `aud`, `iat` or `jti`.

## Alternatives considered

- **`iss` and `aud`** — with one application per deployment there is no value
  to check them against, and adding claims a verifier is told to ignore
  encourages sloppy verification.

## Consequences

`iss` and `aud` become worthwhile if a single deployment ever serves multiple
audiences.
