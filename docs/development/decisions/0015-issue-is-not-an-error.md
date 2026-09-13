# `Issue` is not an `error`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation

## Context

`Issue` describes a failed rule; it is data, not a failure.

## Decision

`Issue` does not implement the `error` interface. Two concrete reasons:

1. Nothing treats an individual issue as an error. Code works with `Issues` or
   with `ValidationError`.
2. `Issue` holds a `map` in `details`, so it is not comparable. If it
   implemented `error`, `errors.Is(err, IssueTooShort(5, UnitCodePoint))` would
   return `false` silently — `errors.Is` skips comparison for non-comparable
   targets. A check that always fails without complaining is worse than no
   check.
