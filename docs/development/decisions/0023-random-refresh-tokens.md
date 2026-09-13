# Refresh tokens are random bytes, not JWTs

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** authentication

## Context

A refresh token is looked up in the database on every use, so it needs no
self-contained content. Its only job is to be unguessable.

## Decision

Refresh tokens are 32 bytes from `crypto/rand`, base64url encoded.

## Alternatives considered

- **A JWT** — would be larger, would invite clients to read its claims, and
  would carry no benefit.
