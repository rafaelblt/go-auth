# `Accumulator.Add` returns nothing

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation
- **Related:** [0014](0014-input-constructors-return-issues.md)

## Context

A request carries several fields, and all of their failures are reported
together.

## Decision

```go
acc.Add(FieldUsername, usernameIssues)
acc.Add(FieldPassword, passwordIssues)
if err := acc.Err(); err != nil { ... }
```

Failures are collected by `Add` and surfaced once by `Err()`.

## Alternatives considered

- **`Add` returning an error** — would put an `if err != nil` between every
  pair of lines, tripling the size of the validation block for no benefit.
