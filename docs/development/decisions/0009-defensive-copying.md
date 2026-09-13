# Defensive copying on reference-typed fields

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain, validation

## Context

Copying a struct copies a slice header or map header, not the underlying data.
A caller retaining the slice it passed to a constructor could mutate a stored
hash after the fact — a silent, hard-to-trace corruption of exactly the data
that must not change.

## Decision

Reference-typed fields are copied on the way in and on the way out:
`RefreshTokenHash` clones on construction and on read, `Issue.Details()`
returns `maps.Clone`, and pointer fields go through `shared.ClonePtr`.
