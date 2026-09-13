# Generic `env[T]`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

Environment variables are resolved the same way and differ in the type they
parse into.

## Decision

One generic type with a `Parser func(string) (T, error)` handles every
variable.

## Alternatives considered

- **One type per value type** (`stringVar`, `intVar` and so on) — the earlier
  design. Abandoned because any change to the shared struct required updating
  every concrete type in lockstep.
