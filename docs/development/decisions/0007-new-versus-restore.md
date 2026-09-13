# `New` versus `Restore`

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain
- **Related:** [0010](0010-username-rules.md), [0011](0011-password-input-rules.md)

## Context

Entities are built both from new input and from rows already stored. Input
policy is a boundary decision that can legitimately be tightened later —
raising the minimum username length, say.

## Decision

`NewX` applies input policy *and* structural invariants. `RestoreX` applies
only structural invariants.

Structural invariants — non-nil ID, non-empty hash — are true of any valid
entity and are enforced in both.

## Alternatives considered

- **`Restore` enforcing the same rules as `New`** — tightening a rule would
  make existing rows unreadable and turn a policy change into a data migration.
