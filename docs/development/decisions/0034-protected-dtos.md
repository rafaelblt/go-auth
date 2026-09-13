# Protected DTOs

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** api

## Context

`internal/usecase/dtos.go` uses exported fields, breaking the
[protected-struct convention](../../architecture/conventions.md#protected-struct)
used everywhere else. That is an inconsistency rather than a principle.

## Alternatives considered

- **Keep exported fields** — DTOs are short-lived, so the risk is low.
- **Apply the protected-struct convention** — removes the inconsistency.
