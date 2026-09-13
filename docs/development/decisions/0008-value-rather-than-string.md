# `Value()` rather than `String()` for sensitive types

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain

## Context

A type with a `String()` method is picked up by `fmt` and `slog`, so a struct
containing it can print its own contents through a `%v` nobody audited.

## Decision

`password.Plain`, `password.Hashed`, `session.AccessToken` and
`session.RefreshTokenHash` expose `Value()`, which has to be called on purpose.

`Username` and `Status` keep `String()` — they are not sensitive and printing
them is useful.
