# `Issues` rather than `error` from input constructors

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation, domain
- **Related:** [0019](0019-accumulator-add-returns-nothing.md)

## Context

Input constructors (`user.NewUsername`, `password.NewPlain`) validate data a
user sent. When they fail, the failure is always a set of broken rules.

## Decision

```go
func NewUsername(value string) (Username, validation.Issues)
```

The return type is the contract: a failure here is always a structured set of
issues, never some other error. `Accumulator.Add` can therefore take `Issues`
directly.

Value objects built from internally generated data (`password.Hashed`,
`session.AccessToken`) return plain `error` instead. Their failures are
programming errors, not user input problems, and there is nothing structured to
report.

## Alternatives considered

- **Returning `error`** — every consumer would have had to handle a possibility
  that never occurs in practice: an `errors.As` to recover the issues, plus a
  branch for the non-validation error that cannot happen.
