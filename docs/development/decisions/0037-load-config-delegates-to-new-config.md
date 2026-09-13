# `LoadConfig` delegates to `NewConfig`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

`Config` can be built from environment variables (`LoadConfig`) or from
parameters passed directly (`NewConfig`).

## Decision

Environment loading resolves variables, then calls `NewConfig` rather than
building `Config` directly, so struct construction lives in one place and
adding a field means changing one function.

`NewConfig` returns an error, but after successful environment resolution it
cannot legitimately fail — the parsers already guarantee well-formed values. A
failure there means a programming error (a new required field, a changed
format), so `LoadConfig` panics rather than starting with a config it cannot
justify.
