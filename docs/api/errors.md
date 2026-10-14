# Error model

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

## Codes

| Code | Status | Endpoints | Meaning |
|---|---|---|---|
| `invalid_json_body` | 400 | every `POST` | The body is not a JSON object with fields of the expected types ([Request bodies](reference.md#request-bodies)) |
| `invalid_credentials` | 401 | login, change password | The username or the password was rejected, for any reason |
| `invalid_token` | 401 | refresh, verify | The token cannot be used, for any reason |
| `route_not_found` | 404 | any unknown path | No endpoint has this path |
| `method_not_allowed` | 405 | every endpoint | Wrong method; `Allow` lists the accepted ones |
| `username_already_exists` | 409 | register | The username is taken |
| `request_body_too_large` | 413 | every `POST` | The body's JSON value runs past 64 KiB |
| `unsupported_media_type` | 415 | every `POST` | The `Content-Type` is not `application/json` |
| `validation_failed` | 422 | register, change password | The input broke a rule; `fields` says which |
| `too_many_requests` | 429 | every `POST` but verify | Rate limiting is on and the address used up its allowance; `Retry-After` gives the seconds to wait |
| `internal_server_error` | 500 | all | An unexpected failure |

## Validation errors

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
(`"josé"` is 4), or `byte`, UTF-8 bytes (`"josé"` is 5). JavaScript's
`String.prototype.length` counts neither: `"😀"` is 2 there, where the server
counts 1 code point and 4 bytes. In a client, count code points with
`[...str].length`, and bytes with `new TextEncoder().encode(str).length`.

## Deliberately vague errors

Four endpoints say less than they know, on purpose. The server logs which
failure it was, as a `reason` on the request's log line
([Outcomes](../logging.md#outcomes)), so nothing is lost for whoever runs the
service.

- **Login** answers `invalid_credentials` to an unknown username, a user with
  no password, a wrong password, and a value that breaks the registration
  rules alike, and a missing account costs a bcrypt comparison like a wrong
  password does. So login tells an unknown username from a wrong password
  neither by its answer nor, within a database round trip, by its time
  ([unless `BCRYPT_COST` changed](../limitations.md#login-timing-after-a-bcrypt_cost-change)).
  Registration, though, does reveal whether a username exists
  ([Limitations](../limitations.md#registration-reveals-whether-a-username-exists)).
- **Change password** checks the username and the current password as login
  does, and answers the same way. Only the new password gets a `422`, which
  depends on nothing but the two passwords sent.
- **Refresh** answers `invalid_token` to a token that never existed, one that
  expired, one already used, and one whose session was revoked, so the answer
  tells someone holding a stolen token nothing about its state.
- **Verify** answers `invalid_token` to a malformed, a forged and an expired
  token alike: a service has one thing to do with any of them, turn the request
  away.

## Handling errors in a client

```
400  →  bug in your client: fix the request
401  →  login: "wrong username or password"
        change password: "wrong username or current password"
        refresh: discard the token and send the user to log in
        verify: turn away the request that carried the token
404, 405, 413, 415  →  bug in your client: the path, method, body size or Content-Type
409  →  ask for another username
422  →  map error.fields[].field to your form fields and show a message for each
429  →  wait Retry-After seconds, then send the same request again
500  →  retry with backoff; if it persists, the service's logs say why
```
