# Error model

Every error response has the same shape: an `error` object with a `code` and
a `message`. A validation error adds `fields`, which names each field that
failed. A client reads `error.code` first, whatever the status.

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
| `message` | Text for humans, set by the error's class rather than by its code. Do not parse it, and reword it before showing it to users. |
| `fields` | Present only when `code` is `validation_failed`: every field that failed. See [Validation errors](#validation-errors). |

Because the message comes from the class, several codes share one text: every
`401` reads `"Not authorized."`, whether the credentials were wrong or the
refresh token was spent. Only `code` says what happened.

## Validation errors

A `422` has the code `validation_failed`, and its `fields` reports every field
that failed, in one response.

```json
{
  "error": {
    "code": "validation_failed",
    "message": "The input failed validation.",
    "fields": [
      {
        "field": "username",
        "code": "too_short",
        "details": { "min": 3, "unit": "code_point" }
      },
      {
        "field": "password",
        "code": "too_long",
        "details": { "max": 72, "unit": "byte" }
      }
    ]
  }
}
```

| Field | Meaning |
|---|---|
| `field` | The request field that failed, named as in the JSON body |
| `code` | The rule that was broken, from the [field codes](#field-codes) |
| `details` | The rule's parameters, shaped by its code |

A field can appear more than once: a username that is too long and contains a
forbidden character produces two entries.

`details` is an empty object when the rule has no parameters.

There is no separate code for a missing or empty field. A missing `username`
is decoded as `""`, which breaks the minimum length, and is reported as
`too_short` with `min: 3`.

### Length units

`details.unit` says how a length was counted, which matters for input that is
not ASCII:

| Unit | Counted as |
|---|---|
| `code_point` | Unicode code points: `"josé"` is 4 |
| `byte` | UTF-8 bytes: `"josé"` is 5 |

Every length rule names its unit, so a client can apply the same rule and get
the same answer. JavaScript's `String.prototype.length` counts UTF-16 code
units, which matches neither: `"josé"` is 4 there, but `"😀"` is 2, where the
server counts 1 code point and 4 bytes. To apply these rules in a client,
count code points with `[...str].length`, and bytes with
`new TextEncoder().encode(str).length`.

## Status codes

| Status | When |
|---|---|
| `200 OK` | Success, except for registration |
| `201 Created` | A user was registered |
| `400 Bad Request` | The body is not a JSON object with fields of the expected types |
| `401 Unauthorized` | Wrong credentials, or an unusable refresh token |
| `404 Not Found` | Unknown path |
| `405 Method Not Allowed` | Known path, wrong method; the `Allow` header lists the accepted methods |
| `409 Conflict` | The username is taken |
| `413 Request Entity Too Large` | A `POST` body whose JSON value runs past 64 KiB |
| `415 Unsupported Media Type` | A `POST` whose `Content-Type` is not `application/json` |
| `422 Unprocessable Entity` | The input failed validation |
| `429 Too Many Requests` | Rate limiting is on and the address used up a `POST` endpoint's allowance; `Retry-After` gives the seconds to wait |
| `500 Internal Server Error` | An unexpected failure |

## Code catalog

### Error codes

The values of `error.code`.

| Code | Status | Endpoint | Meaning |
|---|---|---|---|
| `invalid_json_body` | 400 | every `POST` | The body could not be decoded: not JSON, not valid UTF-8, not an object, or a field of the wrong type |
| `route_not_found` | 404 | any unknown path | No endpoint has this path |
| `method_not_allowed` | 405 | every endpoint | The endpoint does not accept this method; `Allow` lists the ones it does |
| `request_body_too_large` | 413 | every `POST` | The body's JSON value runs past 64 KiB ([Request bodies](reference.md#request-bodies)) |
| `unsupported_media_type` | 415 | every `POST` | The request's `Content-Type` is not `application/json` |
| `too_many_requests` | 429 | every `POST` | Rate limiting is on and the address made too many requests to the endpoint; `Retry-After` gives the seconds to wait |
| `invalid_credentials` | 401 | `/v1/auth/login` | Login failed, for any reason |
| `invalid_token` | 401 | `/v1/auth/refresh` | The refresh token cannot be used, for any reason |
| `username_already_exists` | 409 | `/v1/auth/register` | The username is taken |
| `validation_failed` | 422 | `/v1/auth/register` | The input failed validation; `fields` says how |
| `internal_server_error` | 500 | all | An unexpected failure |

### Field codes

The values of `error.fields[].code`.

| Code | `details` | Meaning |
|---|---|---|
| `too_short` | `{ "min": int, "unit": string }` | Shorter than the minimum length |
| `too_long` | `{ "max": int, "unit": string }` | Longer than the maximum length |
| `invalid_characters` | `{}` | Contains a character outside the allowed set |

`invalid_characters` does not say *which* character failed.

These three are the complete set a response can carry. `internal/validation`
defines others — `required`, `not_positive`, `not_allowed` — but they belong to
the configuration check and appear only in the startup error, never in a
response; see [Validation](../architecture/domain/validation.md#validators).

### Messages

For completeness, since the message is set by the error's class:

| Status | `message` |
|---|---|
| `400` | `Request body is not valid JSON.` |
| `401` | `Not authorized.` |
| `404` | `Route not found.` |
| `405` | `Method not allowed for this route.` |
| `409` | `A conflict error occurred.` |
| `413` | `Request body is too large.` |
| `415` | `Content-Type must be application/json.` |
| `422` | `The input failed validation.` |
| `429` | `Too many requests.` |
| `500` | `An internal error occurred.` |

The `401` and `409` texts come from the error's `kind`, so they cover every code
of that class: a wrong password and a spent refresh token read the same. The
`400`, `404`, `405`, `413`, `415`, `422`, `429` and `500` texts belong to one
code each.

Do not depend on any of them. They are listed so that it is obvious one text
serves several codes, which is the reason to branch on `code`.

## Deliberately vague errors

Two endpoints say less than they know, on purpose.

**Login** answers `invalid_credentials` to a malformed username, an unknown
user, a user with no password, and a wrong password alike. Saying which one
failed would let anyone check whether an account exists. For the same reason,
login answers a malformed username with `401`, not `422`, and takes as long
for an unknown user as for a wrong password.

**Refresh** answers `invalid_token` to a token that never existed, one that
expired, one already used, and one whose session was revoked. Someone holding
a stolen token learns nothing about its state, and in particular not whether
reuse detection has already fired.

In both cases the server logs which of the failures it was, under `reason`, on
the line tagged with the request ID. What the client is not told is still
there for whoever runs the service.

These codes cannot drive detailed error messages. Treat every `401` from
`/v1/auth/refresh` the same way: discard the stored token and send the user to
log in.

## Request IDs

The service gives every request a `request_id` (a UUID) and adds it to every
log line the request produces. The ID is not returned to the client, in the
body or in a header. To find a request in the logs, search by time, path and
status.

## Handling errors in a client

```
400  →  bug in your client: fix the request
401  →  on /login: show "wrong username or password"
        on /refresh: discard the token and send the user to log in
409  →  ask for another username
413  →  bug in your client: the body is too large
415  →  bug in your client: send Content-Type: application/json
422  →  map error.fields[].field to your form fields and show a message for each
429  →  wait Retry-After seconds, then send the same request again (on /refresh,
        the same token: it was not spent)
404  →  bug in your client: the path is wrong
405  →  bug in your client: use the method in the Allow header
500  →  retry with backoff; if it persists, check the service's logs
```

Branch on `code`, never on `message`. Treat a `code` you do not recognise as a
generic failure: new codes can be added within a major version.
