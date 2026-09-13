# JWKS is a plain handler

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** api
- **Related:** [0031](0031-use-case-adapter.md)

## Context

JWKS is the one endpoint with no business logic: it reads keys from a provider
and reshapes them into JWK objects.

## Decision

The JWKS endpoint is served by a plain handler, not by a use case behind the
[adapter](0031-use-case-adapter.md).

## Alternatives considered

- **A use case** — would add a layer that only forwarded data.
