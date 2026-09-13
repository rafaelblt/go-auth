# Ports and adapters

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** architecture

## Context

The use cases need a database, a password hasher and a token signer, but the
logic they implement does not depend on which technology provides them.

## Decision

The use case layer depends on interfaces in `internal/port`, never on
PostgreSQL, bcrypt or `net/http`. Infrastructure implements those interfaces
and is injected at startup.

## Consequences

The immediate payoff is test speed: every use case is tested against in-memory
fakes, so the logic suite runs in milliseconds. The longer-term payoff is that
swapping bcrypt for argon2, or PostgreSQL for something else, is a new package
implementing an existing interface rather than a rewrite.
