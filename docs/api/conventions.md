# API conventions

What every endpoint shares. Each endpoint is described with the rest of its
topic, and [Endpoints](../system.md#endpoints) lists them. The same contract,
machine-readable, is [openapi.yaml](openapi.yaml) (OpenAPI 3.1); a change to the
contract updates both.

## Versioning

The auth endpoints start with `/v1`. The JWKS has no version, because
[RFC 8615](https://www.rfc-editor.org/rfc/rfc8615) fixes its path.

## Content type

The `POST` endpoints require `Content-Type: application/json`, matched
regardless of case, with any parameters (`; charset=utf-8`) ignored. Anything
else, a missing header included, is answered `415 unsupported_media_type`
before the body is read, and is never counted by the
[rate limit](#rate-limiting). `curl -d` sends a form type, so add
`-H 'Content-Type: application/json'` or use `--json`.

Every response body is JSON, except the redirects of
[Methods and routes](#methods-and-routes), the empty body of a `500` whose
response could not be encoded, and the `204` of
[logout](../sessions.md#logging-out), which has no body and no `Content-Type`.
`go-auth` sends no CORS headers.

**Why:** a browser sends a cross-site `POST` without asking the site first when
its body is a form, URL-encoded or `text/plain`; for any other type,
`application/json` included, it first sends a CORS preflight, which `go-auth`
never grants. Without the requirement, a page on any site could send logins and
registrations through its visitors' browsers, each from its own address, past
any per-address limit, and counting those requests would let the page spend its
visitors' allowance. A reverse proxy that answers CORS preflights for these
endpoints, for origins the operator does not control, undoes this.

## Methods and routes

The auth endpoints accept `POST`, and the JWKS `GET` and `HEAD`. Another method
on a known path is `405 method_not_allowed`, with an `Allow` header naming the
accepted ones; an unknown path is `404 route_not_found`. A path not in
canonical form (`//v1/auth/login`, or one with `..` segments) gets a `307`
redirect to its cleaned form from Go's `http.ServeMux`: a short HTML body for a
`GET`, an empty one otherwise.

## Request bodies

The body of a `POST` is one JSON object. An empty body, malformed or truncated
JSON, a value that is not an object (`[]`, `"text"`), a field of the wrong type
(`{"username": 1}`), and invalid UTF-8 anywhere in the object, in a field that
is ignored too, are all `400 invalid_json_body`. Invalid UTF-8 is rejected
rather than replaced because every invalid byte would become the same `U+FFFD`,
so a password sent in Latin-1 would be stored that way: `senhaçã` and
`senhaõé` would be one password.

A body is decoded up to its first 64 KiB (65,536 bytes): one whose first JSON
value does not end within them is `413 request_body_too_large`. The limit is not
a setting.

**Unknown fields are ignored**, which is what lets a field be added without
breaking clients. A misspelled field is therefore not reported as such: the
field it was meant to be arrives empty, and is reported instead
(`{"usernme": "alice"}` at registration is `too_short` on `username`). A missing
field is read as `""`.

The decoder also tolerates the following. Send one well-formed object, in valid
UTF-8, with each field once and in lower case, and do not depend on these:

- a body of `null` reads as `{}`;
- a field set to `null` is read as `""`, like a missing one;
- anything after the first JSON value is ignored;
- field names match regardless of case, and a field sent twice keeps the last
  value;
- an escaped lone surrogate such as `\ud800` becomes `U+FFFD`, which in a
  password has consequences
  ([Limitations](../accounts.md#a-lone-surrogate-in-a-password-becomes-ufffd)).

## Timestamps

RFC 3339, in UTC, to the whole second (`2026-09-08T12:00:00Z`). A fraction of a
second is dropped, never rounded up, so an `expires_at` is never later than the
real expiry.

## Errors

Every error response has the same shape: an `error` object with a `code` and a
`message`, and, for a validation error, `fields`.

```json
{
  "error": {
    "code": "invalid_credentials",
    "message": "Not authorized."
  }
}
```

| Field | Meaning |
|---|---|
| `code` | A stable identifier for the failure. Branch on this. |
| `message` | Text for humans, set by the status rather than by the code: every `401` reads `Not authorized.`, whether the password was wrong or the refresh token spent. Do not parse it. |
| `fields` | Only when `code` is `validation_failed`: every rule that failed. |

New codes can be added within a major version, so treat a code you do not
recognise as a generic failure.

Each endpoint lists its own errors. Besides those, every `POST` endpoint can
answer `400 invalid_json_body`, `413 request_body_too_large`,
`415 unsupported_media_type` and `500 internal_server_error`, a rate-limited one
`429 too_many_requests`, and any path `404 route_not_found` or
`405 method_not_allowed`.

### Codes

| Code | Status | Meaning |
|---|---|---|
| `invalid_json_body` | `400` | The body is not a JSON object with fields of the expected types ([Request bodies](#request-bodies)) |
| `invalid_credentials` | `401` | The username or the password was rejected, for any reason ([Checking credentials](../accounts.md#checking-credentials)) |
| `invalid_token` | `401` | The token cannot be used, for any reason |
| `route_not_found` | `404` | No endpoint has this path |
| `method_not_allowed` | `405` | Wrong method; `Allow` lists the accepted ones |
| `username_already_exists` | `409` | The username is taken |
| `request_body_too_large` | `413` | The body's JSON value runs past the [limit](#request-bodies) |
| `unsupported_media_type` | `415` | The `Content-Type` is not `application/json` ([Content type](#content-type)) |
| `validation_failed` | `422` | The input broke a rule; `fields` says which |
| `too_many_requests` | `429` | The client address used up its allowance; `Retry-After` gives the seconds to wait ([Rate limiting](#rate-limiting)) |
| `internal_server_error` | `500` | An unexpected failure |

### Validation errors

```json
{
  "error": {
    "code": "validation_failed",
    "message": "The input failed validation.",
    "fields": [
      { "field": "username", "code": "too_short", "details": { "min": 3, "unit": "code_point" } },
      { "field": "password", "code": "too_long", "details": { "max": 72, "unit": "byte" } }
    ]
  }
}
```

`field` names the request field as in the JSON body, and can appear more than
once: a username that is too long and has a forbidden character gives two
entries. `code` is one of these, the complete set a response can carry:

| Code | `details` | Meaning |
|---|---|---|
| `too_short` | `{ "min": int, "unit": string }` | Shorter than the minimum length |
| `too_long` | `{ "max": int, "unit": string }` | Longer than the maximum length |
| `invalid_characters` | `{}` | A character outside the allowed set; which one is not said |
| `same_as_current` | `{}` | On change password: `new_password` is `current_password` |

There is no code for a missing field: it is read as `""`, which is `too_short`.

`details.unit` says how a length was counted: `code_point`, Unicode code points
(`"josé"` is 4), or `byte`, UTF-8 bytes (`"josé"` is 5). A length always carries
its unit because clients count differently: JavaScript's
`String.prototype.length` counts neither, so `"😀"` is 2 there, where the server
counts 1 code point and 4 bytes. In a client, count code points with
`[...str].length`, and bytes with `new TextEncoder().encode(str).length`.

### Deliberately vague answers

A `401` says less than the server knows, on purpose: one code covers every
failure of its kind, and the log says which one it was, as the request's
`reason` ([Outcomes](../logging.md#outcomes)), so nothing is lost for whoever
runs the service. Each endpoint says what its code covers and what it keeps
back.

### Handling errors in a client

```
400  →  bug in your client: fix the request
401  →  what the endpoint's errors table says
404, 405, 413, 415  →  bug in your client: the path, method, body size or Content-Type
409  →  what the endpoint's errors table says
422  →  map error.fields[].field to your form fields and show a message for each
429  →  wait Retry-After seconds, then send the same request again
500  →  retry with backoff; if it persists, the service's logs say why
```

## Rate limiting

Off unless [`RATE_LIMIT`](../configuration.md#rate_limit) chooses a level, which
sets the allowance of each limited endpoint; the table there lists them. The
allowances are per client address and per endpoint: a login does not spend a
client's refreshes. A request past its allowance is answered
`429 too_many_requests`, with a `Retry-After` header in seconds.

- **The limit is decided before the body is read**, so a refused request read
  and ran nothing: the same request can be sent again after `Retry-After`.
- **Requests answered `415` are never counted** ([why](#content-type)).
- **Every other request counts**, whatever its answer: the decision is taken
  before it is known.
- **A failure of the limiter is a `500`**, and the request does not run: failing
  open would remove the limit without anyone noticing.

### Which address counts

The client address is the TCP peer, unless the peer is one of the
[`TRUSTED_PROXIES`](../configuration.md#trusted_proxies). Then `X-Forwarded-For`
is read from the right, skipping each trusted proxy, and the client is the first
address that is not one, or the leftmost when all are. An empty entry, or one
that is not a bare address (`ip:port` included), stops the walk at the last good
address. Reading from the right is what keeps a client from choosing its
address: the leftmost entries are whatever it sent, and only those a trusted
proxy appended can be believed.

Addresses are normalised alike: an IPv4-mapped IPv6 address counts as IPv4, and
a zone is dropped. An IPv6 client counts as its /64, since one host normally
holds a whole /64 and could otherwise rotate through it. The address counted is
logged with each refusal ([Outcomes](../logging.md#outcomes)); the request lines
keep the peer.

## Limitations

### Rate limiting is per address and per process

- **Per address, not per account.** Guesses against one account spread over
  many addresses are not limited, whether a proxy or the service limits them.
- **An address can be many users.** Clients behind one NAT share an allowance,
  and so does an IPv6 /64; `relaxed` exists for that. The users of an
  application server that calls the service for them share its allowance too,
  unless it forwards their addresses.
- **In memory.** A restart restores every allowance, and each instance counts
  on its own: N instances allow N times the limit.
- **Off by default**, which is [deprecated](../configuration.md#rate_limit).
