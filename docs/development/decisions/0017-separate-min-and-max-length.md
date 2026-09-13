# Separate `MinLength` and `MaxLength`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation

## Context

Length rules need a lower bound, an upper bound, or both.

## Decision

Length is checked by two validators, `MinLength` and `MaxLength`. This also
keeps the mapping one validator, one issue code.

## Alternatives considered

- **A single `Length(min, max int)`** — more compact, but two adjacent `int`
  parameters invite swapping, and `Length(32, 3)` is an impossible rule that
  nothing would reject.
