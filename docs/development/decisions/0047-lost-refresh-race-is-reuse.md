# A refresh that loses the race for its token is reuse

- **Status:** Accepted
- **Date:** 2026-09-13
- **Areas:** authentication
- **Related:** [0046](0046-refresh-token-use-guarded-at-write.md), [Reuse detection](../../architecture/usecases/refresh.md#reuse-detection)

## Context

With [the use guarded at write](0046-refresh-token-use-guarded-at-write.md),
refresh can learn that a token was already spent in two places: from
`RefreshToken.Use`, when the token was spent before it was read, and from
`RefreshTokenWriter.MarkUsed`, when a concurrent request spent it between the
read and the write. [Reuse detection](../../architecture/usecases/refresh.md#reuse-detection) revokes the
session in the first case. The second case needed its own answer.

## Decision

Both cases are reuse. When `MarkUsed` returns `session.ErrTokenAlreadyUsed`,
the rotation transaction rolls back, refresh revokes the session and returns
`ErrTokenAlreadyUsed`, exactly as for a token found spent on read.

Two requests presenting the same token at the same instant are the situation
reuse detection describes: two parties hold the token, and the service cannot tell which
is which. Arriving at the same moment instead of one after the other does not
make either more trustworthy. If the attacker wins the race, revoking is the
only thing that locks them out.

## Alternatives considered

- **Answer `401` without revoking** — spares a client that refreshes from
  several tabs at once. But if the legitimate client loses and does not retry
  with the same token, an attacker who won keeps a rotating credential, which
  is what reuse detection exists to prevent.
- **Answer `500`** — the behaviour before 0046, by accident. It hides a
  security signal behind an infrastructure error.

## Consequences

A client that sends concurrent refreshes with one token is logged out, in
every tab. Clients must serialize refreshes. Whether the service should
tolerate legitimate duplicates is left open; see
[Limitations](../../limitations.md#a-repeated-refresh-logs-the-user-out).
