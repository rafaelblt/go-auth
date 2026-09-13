# Startup fails on a schema mismatch

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** configuration

## Context

The migration version applied to the database can differ from the latest one
embedded in the binary.

## Decision

If the applied migration version differs from the latest embedded one, in
either direction, the service exits.

Failing at startup turns a partial outage into a deploy that visibly does not
roll out.

## Alternatives considered

- **Starting anyway** — a binary expecting a column that does not exist would
  fail one request at a time, in production, as a `500`.
