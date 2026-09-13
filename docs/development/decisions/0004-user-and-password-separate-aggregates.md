# `user` and `password` as separate aggregates

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain

## Context

Authentication methods other than passwords may be added later. Originally
there was a generic `Credential` entity meant to cover them.

## Decision

`Password` is a concrete entity, in its own aggregate, separate from `User`.

A concrete `Password` says exactly what it is. Keeping it a separate entity
from `User` still expresses that a user *may* have a password, leaving room to
add other credential types as their own aggregates when they actually arrive.

## Alternatives considered

- **A generic `Credential` entity** — the original design. It could not be
  documented honestly: its fields and semantics were speculation about
  credential types that did not exist.
