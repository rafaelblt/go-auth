# `kind`-based status mapping

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** api
- **Related:** [0002](0002-ports-and-adapters.md), [0028](0028-uniform-auth-errors.md)

## Context

Use case failures end up as HTTP responses, but the use case layer must not
know about HTTP.

## Decision

`UseCaseError` carries a `kind` (`conflict`, `unauthorized`) rather than an
HTTP status. The API layer owns the mapping from kind to status.

## Alternatives considered

- **An HTTP status on the error** — would put `net/http` concerns into
  `internal/usecase`.

## Consequences

The kinds do line up closely with HTTP status classes, which is an acknowledged
bias — but the indirection is what keeps `internal/usecase` free of `net/http`.
