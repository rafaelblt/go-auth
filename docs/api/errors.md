# Error model

Every error response has the same shape: an `error` object with a `code` and
a `message`. A validation error adds `fields`, which names each field that
failed. A client reads `error.code` first, whatever the status. The one
exception is the plain-text `404` and `405` of an unknown route; see
[Status codes](#status-codes).

```json
{
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Not authorized."
  }
}
```

| Field | Meaning |
|---|---|
| `code` | A stable identifier for the failure. Branch on this. |
| `message` | Text for humans, set by the error's class rather than by its code. Do not parse it, and reword it before showing it to users. |
| `fields` | Present only when `code` is `VALIDATION_FAILED`: every field that failed. See [Validation errors](#validation-errors). |

Because the message comes from the class, several codes share one text: every
`401` reads `"Not authorized."`, whether the credentials were wrong or the
refresh token was spent. Only `code` says what happened.

## Validation errors

A `422` has the code `VALIDATION_FAILED`, and its `fields` reports every field
that failed, in one response.

```json
{
  "error": {
    "code": "VALIDATION_FAILED",
    "message": "The input failed validation.",
    "fields": [
      {
        "field": "username",
        "code": "TOO_SHORT",
        "details": { "min": 3, "unit": "code_point" }
      },
      {
        "field": "password",
        "code": "TOO_LONG",
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
`TOO_SHORT` with `min: 3`.

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
| `404 Not Found` | Unknown path. **Plain text body**, not JSON |
| `405 Method Not Allowed` | Known path, wrong method. **Plain text body**, not JSON |
| `409 Conflict` | The username is taken |
| `413 Request Entity Too Large` | A `POST` body whose JSON value runs past 64 KiB |
| `415 Unsupported Media Type` | A `POST` whose `Content-Type` is not `application/json` |
| `422 Unprocessable Entity` | The input failed validation |
| `500 Internal Server Error` | An unexpected failure |

## Code catalog

### Error codes

The values of `error.code`.

| Code | Status | Endpoint | Meaning |
|---|---|---|---|
| `INVALID_JSON_BODY` | 400 | every `POST` | The body could not be decoded: not JSON, not valid UTF-8, not an object, or a field of the wrong type |
| `REQUEST_BODY_TOO_LARGE` | 413 | every `POST` | The body's JSON value runs past 64 KiB ([Request bodies](reference.md#request-bodies)) |
| `UNSUPPORTED_MEDIA_TYPE` | 415 | every `POST` | The request's `Content-Type` is not `application/json` |
| `INVALID_CREDENTIALS` | 401 | `/v1/auth/login` | Login failed, for any reason |
| `INVALID_TOKEN` | 401 | `/v1/auth/refresh` | The refresh token cannot be used, for any reason |
| `USERNAME_ALREADY_EXISTS` | 409 | `/v1/auth/register` | The username is taken |
| `VALIDATION_FAILED` | 422 | `/v1/auth/register` | The input failed validation; `fields` says how |
| `INTERNAL_SERVER_ERROR` | 500 | all | An unexpected failure |

### Field codes

The values of `error.fields[].code`.

| Code | `details` | Meaning |
|---|---|---|
| `TOO_SHORT` | `{ "min": int, "unit": string }` | Shorter than the minimum length |
| `TOO_LONG` | `{ "max": int, "unit": string }` | Longer than the maximum length |
| `INVALID_CHARACTERS` | `{}` | Contains a character outside the allowed set |

`INVALID_CHARACTERS` does not say *which* character failed.

These three are the complete set a response can carry. `internal/validation`
defines others — `REQUIRED`, `NOT_POSITIVE`, `NOT_ALLOWED` — but they belong to
the configuration check and appear only in the startup error, never in a
response; see [Validation](../architecture/domain/validation.md#validators).

### Messages

For completeness, since the message is set by the error's class:

| Status | `message` |
|---|---|
| `400` | `Request body is not valid JSON.` |
| `401` | `Not authorized.` |
| `409` | `A conflict error occurred.` |
| `413` | `Request body is too large.` |
| `415` | `Content-Type must be application/json.` |
| `422` | `The input failed validation.` |
| `500` | `An internal error occurred.` |

The `401` and `409` texts come from the error's `kind`, so they cover every code
of that class: a wrong password and a spent refresh token read the same. The
`400`, `413`, `415`, `422` and `500` texts belong to one code each.

Do not depend on any of them. They are listed so that it is obvious one text
serves several codes, which is the reason to branch on `code`.

## Deliberately vague errors

Two endpoints say less than they know, on purpose.

**Login** answers `INVALID_CREDENTIALS` to a malformed username, an unknown
user, a user with no password, and a wrong password alike. Saying which one
failed would let anyone check whether an account exists. For the same reason,
login answers a malformed username with `401`, not `422`, and takes as long
for an unknown user as for a wrong password.

**Refresh** answers `INVALID_TOKEN` to a token that never existed, one that
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
404  →  bug in your client; the body is plain text
405  →  bug in your client; the body is plain text
500  →  retry with backoff; if it persists, check the service's logs
```

Branch on `code`, never on `message`. Treat a `code` you do not recognise as a
generic failure: new codes can be added within a major version.
