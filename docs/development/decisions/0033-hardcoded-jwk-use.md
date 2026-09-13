# `use: "sig"` is hard-coded

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** api

## Context

Each JWK carries a `use` value stating the key's purpose. `port.PublicKey` has
no `use` field, and there is currently exactly one purpose.

## Decision

The JWKS handler hard-codes `use: "sig"`.

## Alternatives considered

- **A `use` field on `port.PublicKey`** — would mean modelling an enumeration
  of key purposes with a single member.

## Consequences

Revisit when a second key purpose exists.
