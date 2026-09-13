# Wait for the readiness log twice

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** testing

## Context

PostgreSQL prints "database system is ready to accept connections" once during
initialisation and again when actually ready.

## Decision

The wait strategy for the test database requires the message twice.

## Alternatives considered

- **Waiting for the first occurrence** — produced flaky connection failures.
