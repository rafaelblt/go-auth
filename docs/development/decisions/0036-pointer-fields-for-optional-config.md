# Pointer fields for optional config

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

For a `string` or `int`, the zero value is also an invalid setting, so a plain
field cannot distinguish "not provided, use the default" from "explicitly set
to something invalid" — and the second case should be an error, not silently
replaced by a default.

## Decision

`ConfigParams` uses `*int` and `*time.Duration` for optional values.
`AutoMigrate` is a plain `bool` because `false` is a valid setting.
