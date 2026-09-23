# Error model

The API has two error shapes: a **single error**, for a request that fails as
a whole, and a **validation error**, for input that fails field by field.

## Single error

Used for every failure except input validation.

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

Because the message comes from the class, several codes share one text: every
`401` reads `"Not authorized."`, whether the credentials were wrong or the
refresh token was spent. Only `code` says what happened.

## Validation error

Used only for `422`. It reports every field that failed, in one response.

```json
{
  "errors": [
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
```

| Field | Meaning |
|---|---|
| `field` | The request field that failed, named as in the JSON body |
| `code` | The rule that was broken |
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
| `200 OK` | Success. Every successful response is `200`, registration included. |
| `400 Bad Request` | The body is not valid JSON |
| `401 Unauthorized` | Wrong credentials, or an unusable refresh token |
| `404 Not Found` | Unknown path. **Plain text body**, not JSON |
| `405 Method Not Allowed` | Known path, wrong method. **Plain text body**, not JSON |
| `409 Conflict` | The username is taken |
| `422 Unprocessable Entity` | The input failed validation |
| `500 Internal Server Error` | An unexpected failure |

## Code catalog

### Single-error codes

| Code | Status | Endpoint | Meaning |
|---|---|---|---|
| `INVALID_JSON_BODY` | 400 | every `POST` | The body could not be decoded as JSON |
| `INVALID_CREDENTIALS` | 401 | `/v1/auth/login` | Login failed, for any reason |
| `INVALID_TOKEN` | 401 | `/v1/auth/refresh` | The refresh token cannot be used, for any reason |
| `USERNAME_ALREADY_EXISTS` | 409 | `/v1/auth/register` | The username is taken |
| `INTERNAL_SERVER_ERROR` | 500 | all | An unexpected failure |

### Validation codes

| Code | `details` | Meaning |
|---|---|---|
| `TOO_SHORT` | `{ "min": int, "unit": string }` | Shorter than the minimum length |
| `TOO_LONG` | `{ "max": int, "unit": string }` | Longer than the maximum length |
| `INVALID_CHARACTERS` | `{}` | Contains a character outside the allowed set |

`INVALID_CHARACTERS` does not say *which* character failed.

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
422  →  map errors[].field to your form fields and show a message for each
404  →  bug in your client; the body is plain text
405  →  bug in your client; the body is plain text
500  →  retry with backoff; if it persists, check the service's logs
```

Branch on `code`, never on `message`. Treat a `code` you do not recognise as a
generic failure: new codes can be added within a major version.
