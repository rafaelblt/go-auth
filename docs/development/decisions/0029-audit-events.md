# Audit events

- **Status:** Proposed
- **Date:** 2026-09-12
- **Areas:** authentication

## Context

Authentication events — logins, refreshes, revocations — are recorded only as
log lines. Structured audit is what a security review of a deployed system will
ask for first.

## Alternatives considered

- **Durable, queryable audit rows** for authentication events.
- **Log lines only** — the current state.
