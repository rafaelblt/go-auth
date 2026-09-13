# A separate HTTP service rather than a library

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** architecture
- **Related:** [0020](0020-two-token-types.md)

## Context

The service is meant to be reusable across projects written in different
languages.

## Decision

`go-auth` is a standalone service with an HTTP+JSON API, which serves anything
that can make a request.

## Alternatives considered

- **A Go library** — would only serve Go projects.

## Consequences

Every auth call costs a network hop. That is why access tokens are verifiable
offline: the hop happens on login and refresh, which are rare, and not on the
requests an application actually serves. See
[Two token types](0020-two-token-types.md).
