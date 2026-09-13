# Presets

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

Some settings are numbers whose meaning depends on specialist knowledge.
`BCRYPT_COST=STRONG` beats `BCRYPT_COST=14` for an operator who has not read
the bcrypt literature.

## Decision

Variables can accept named presets. Preset lookup happens before parsing, so a
preset name never has to be a valid value of the target type.

## Consequences

Presets are case-sensitive. That is a rough edge rather than part of the
decision — `checkPreset` would need only a case fold to fix.
