# The rate limit is decided per client address before the body is read

- **Status:** Accepted
- **Date:** 2026-10-04
- **Related:** [Run it behind a reverse proxy](../../limitations.md#run-it-behind-a-reverse-proxy), [0051](0051-post-endpoints-require-application-json.md), [Rate limiting](../../architecture/http.md#rate-limiting)

## Invariant

An endpoint served by `adaptUseCase` that is rate limited decides its limit
after the `415` check and before it reads the body or runs the use case, and
counts every request that passes the `415`, whatever its answer; `X-Forwarded-For` is read
only when the peer is in `TRUSTED_PROXIES`, from the right, and the client is
the first address that is not a trusted proxy. Moving the check above the
`415` lets any web page spend its visitors' allowance.

## Context

[`RATE_LIMIT`](../../configuration.md#rate_limit) limits the three `POST`
endpoints per client address, and never the JWKS. Where the check sits decides
more than which routes it covers:

- **what a refused request costs the client.** A refresh token can be used
  once. A `429` written after the use case ran would leave the client holding a
  spent token and no answer, and its retry would trip reuse detection and log
  the user out ([client obligations](../../architecture/usecases/refresh.md#client-obligations)).
- **who can spend an allowance.** A browser sends some `POST`s cross-site
  without asking the site first ([0051](0051-post-endpoints-require-application-json.md)).
  Whatever counts those lets a page on any site spend its visitors' allowances.
- **what the service reads before deciding.** A body can be 64 KiB.

The client address has the same kind of question. Behind a reverse proxy the
TCP peer is the proxy, and the client address is in `X-Forwarded-For`, a header
any client can also write.

## Decision

The check runs in `adaptUseCase`, right after the `415` block and before
`http.MaxBytesReader` wraps the body. The adapter serves exactly the three
`POST` endpoints and runs after the `ServeMux`, so the JWKS, a `404` and a
`405` are never limited. Each endpoint is limited only when `NewRouter` passes
it a limit, so an endpoint added later needs its own limit, in `NewRouter` and
in each level's table. Nothing has been read or run when it decides, so a
`429` never spends a refresh token, and the same request can be sent again
after `Retry-After`. After the `415`, only `application/json` requests reach
it, and a browser sends those cross-site only after a CORS preflight the
service never grants.

Every request that passes the `415` counts, whatever its answer: a decision
taken before the use case runs cannot know the answer.

The client address is the TCP peer. Only when the peer is in
[`TRUSTED_PROXIES`](../../configuration.md#trusted_proxies) is
`X-Forwarded-For` read, from the right, skipping trusted addresses; the first
address that is not one is the client. Every address is normalised alike
(IPv4-mapped IPv6 becomes IPv4, a zone is dropped), and an IPv6 client counts
as its /64. The walk is described in
[Client address](../../architecture/http.md#client-address).

A limiter error is answered `500`, and the use case does not run.

## Alternatives considered

- **Before the `415`**: a browser sends `text/plain` and form `POST`s
  cross-site without a preflight ([0051](0051-post-endpoints-require-application-json.md)).
  Counted, they would let any page keep its visitors' address, and everyone
  behind the same NAT or trusted proxy, locked out of login and registration.
  A `415` reads and runs nothing, so counting it protects nothing. It would
  also answer `429` where 0051's invariant requires a `415`.
- **After the decoder**: up to 64 KiB read before deciding, for nothing the
  decision uses.
- **A middleware in front of the router**: it would have to exempt the JWKS
  `GET`, and it would count requests the `ServeMux` answers `404` or `405`,
  the objection 0051 raised for the `415`.
- **A wrapper on each route in `NewRouter`**: it would run before the
  adapter's `415` check, so it would count the requests 0051 answers `415`,
  unless it repeated that check.
- **Counting only failures**: needs the use case's answer, so the `429` would
  come after a refresh token was spent.
- **The leftmost `X-Forwarded-For` entry, or the header whatever the peer**:
  those entries are whatever the client sent, so any client could choose the
  address it is counted as.
- **The full IPv6 address**: one host holds a whole /64, and could rotate
  through 2^64 keys, each with a fresh allowance and its own map entry.
- **Failing open on a limiter error**: an outage of a shared store would
  remove the limit without anyone noticing.

## Consequences

- A `429` is safe to retry, with the same refresh token on `/v1/auth/refresh`.
- An endpoint added later is not limited until `NewRouter` passes it a limit.
- Clients behind one NAT share an allowance, and an IPv6 /64 is one client.
- Behind a proxy that is not in `TRUSTED_PROXIES`, every client is counted as
  the proxy and shares its allowance.
- A malformed or `ip:port` `X-Forwarded-For` entry stops the walk at the last
  trusted address: stricter for the clients behind it, never chosen by them.
- A store shared between processes implements `port.RateLimiter` and changes
  none of this.
