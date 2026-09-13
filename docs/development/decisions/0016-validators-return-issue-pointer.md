# Validators return `*Issue`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation

## Context

A validator checks one rule. Every caller of a validator wants the issue it
reports, not only whether the value passed.

## Decision

```go
type Validator[T any] = func(value T) *Issue
```

`nil` means the value passed.

## Alternatives considered

- **`error`** — would force `errors.As` on every caller that wants the issue,
  and every caller wants the issue.
- **`(T, bool)`** — ambiguous: it is not obvious whether `true` means "valid"
  or "there is an issue".
