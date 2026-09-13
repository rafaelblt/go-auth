# The use case adapter

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** api

## Context

Every endpoint backed by a use case decodes a request, runs the use case,
translates its error, logs, and writes a response.

## Decision

`adaptUseCase` turns any `Execute(ctx, In) (Out, error)` into an
`http.HandlerFunc`, given a decoder, an encoder and a success logger. Decoding,
error translation, logging and writing live in one place, so a new endpoint is
three small functions.

## Alternatives considered

- **A handler per endpoint** — each with its own copy of the error-handling
  logic, and its own opportunity to get it subtly wrong.
