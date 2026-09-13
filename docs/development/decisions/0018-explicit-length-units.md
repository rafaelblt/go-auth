# Explicit length units

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** validation, api

## Context

Go's `len(string)` counts bytes; a JavaScript client's `.length` counts UTF-16
code units; a user counts characters. For `"josé"` those are 5, 4 and 4.

## Decision

Every length rule names its unit, and that unit reaches the client:

```json
{ "code": "TOO_SHORT", "details": { "min": 3, "unit": "code_point" } }
```

Naming the unit makes the rule reproducible on the client.

## Alternatives considered

- **An implicit unit** — in an API where a length rule decides whether a
  request is accepted, leaving the unit implicit guarantees client and server
  eventually disagree.
