# Register output shape

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** api

## Context

Registration returns a full user DTO, so adding a field to `User` automatically
changes the API contract.

## Alternatives considered

- **Keep the full user DTO.**
- **Return only the ID** — decouples the response from the `User` shape.
