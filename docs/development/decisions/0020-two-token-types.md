# Two token types

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication, architecture
- **Related:** [0001](0001-separate-http-service.md)

## Context

A token should be verifiable without a call to this service, and a session
should be revocable. No single token can have both properties.

## Decision

Access tokens are stateless so they can be verified without a call to this
service. Refresh tokens are stateful so they can be revoked.

## Consequences

A revoked session keeps working for up to `ACCESS_TOKEN_TTL`. That is the price
of offline verification, and the reason the default TTL is 30 minutes rather
than a day.
