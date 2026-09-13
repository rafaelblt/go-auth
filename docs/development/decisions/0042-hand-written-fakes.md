# Hand-written fakes

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** testing
- **Related:** [0002](0002-ports-and-adapters.md)

## Context

Use cases are tested against the port interfaces they depend on, which need
test doubles.

## Decision

One hand-written fake per port interface in `internal/testutil/porttest`. They
are small, they read as ordinary Go, the compiler catches interface drift
immediately, and there is no generation step or DSL to learn.

Because they record their parameters, tests can assert interactions as well as
results — that login looked the user up by the *normalised* username, for
instance.

## Alternatives considered

- **A mocking library** — brings a generation step or a DSL to learn.
