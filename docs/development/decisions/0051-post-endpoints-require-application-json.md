# The POST endpoints require Content-Type: application/json

- **Status:** Accepted
- **Date:** 2026-09-24
- **Related:** [Run it behind a reverse proxy](../../limitations.md#run-it-behind-a-reverse-proxy)

## Invariant

Every endpoint served by `adaptUseCase` answers `415` to a request whose
`Content-Type` is not `application/json`, before reading the body, and the
service sends no CORS headers. Removing the check, moving it after the decoder,
or letting another origin's preflight succeed reopens cross-site `POST`s.

## Context

A browser lets any page send a `POST` to another site without asking that site
first, as long as the request is "simple": its `Content-Type` is `text/plain`,
`application/x-www-form-urlencoded` or `multipart/form-data`. For any other
type, `application/json` included, it sends a CORS preflight first, and the
request only if the site allows it. `go-auth` allows none.

The decoders used to read the body as JSON whatever its `Content-Type`, and the
JSON decoder ignores anything after the first value. Two simple requests
therefore reached the use cases with a well-formed body:

- a `<form enctype="text/plain">` with one field, named after the JSON body and
  with no value, which the browser sends as the body followed by `=`, which the
  decoder ignores;
- `fetch(url, {method: "POST", mode: "no-cors", body: "..."})` with a string
  body, which the browser sends as `text/plain;charset=UTF-8`.

The page cannot read the responses, but it does not need to. It spreads login
attempts over its visitors' browsers, each from its own address, past the
per-address rate limit that [Limitations](../../limitations.md#run-it-behind-a-reverse-proxy)
recommends in front of the service; and a registration takes effect without
anyone reading the answer. There is no cookie, so this is not classic CSRF: no
request rides on a visitor's session, and a refresh is of no use to the page,
which would need a token it does not have.

## Decision

`adaptUseCase` checks the `Content-Type` before it calls the decoder. It
accepts the header when `mime.ParseMediaType` parses it without an error as
`application/json`: in any case, and with any parameters, which it ignores,
since RFC 8259 gives `application/json` no `charset`. Anything else, a missing
header included, is answered `415 UNSUPPORTED_MEDIA_TYPE` without the body
being read.

The check is in the adapter because the adapter serves exactly the three
`POST` endpoints, runs after the `ServeMux`, and is what a new use case
endpoint will be served by, without anyone having to remember the check.

## Alternatives considered

- **Rejecting only content after the JSON value**: closes the form, which
  depends on the trailing `=` being ignored, but not the `fetch`, whose body
  can be exact JSON.
- **Doing nothing, and documenting it**: the documentation would have to stop
  presenting the proxy's rate limit as a limit on password guessing, since any
  web page could go around it.
- **A middleware in front of the router**: it would have to exempt the JWKS
  `GET`, and it would answer `415` before the `ServeMux` answers `404` to an
  unknown path.
- **A wrapper on each route in `NewRouter`**: works, but a route added without
  it reopens the hole without anything noticing.

## Consequences

- A client that sends JSON without the header stops working, and `curl -d`
  needs `-H 'Content-Type: application/json'` or `--json`. That is why this
  landed before v1.
- A reverse proxy that answers CORS preflights for these endpoints, for
  origins the operator does not control, undoes the protection.
- The decoder stays as tolerant as it was
  ([Request bodies](../../api/reference.md#request-bodies)); the check does not
  depend on it being strict.
