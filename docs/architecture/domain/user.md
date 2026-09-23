# User

`internal/domain/user`

The account. Created by `/v1/auth/register`, found by username at login.

## User

`internal/domain/user/user.go`

| Field                     | Type        | Notes                |
| ------------------------- | ----------- | -------------------- |
| `id`                      | `ID`        | assigned on creation |
| `username`                | `Username`  | unique, lower case   |
| `status`                  | `Status`    | always `active`      |
| `createdAt` / `updatedAt` | `time.Time` | UTC                  |

`NewUser` requires a username, sets the status to active and `updatedAt` to
`createdAt`. `RestoreUser` rebuilds a user from storage, and also requires an
ID and a status.

## Username

`internal/domain/user/username.go`

| Rule               | Value                       |
| ------------------ | --------------------------- |
| Minimum length     | 3 code points               |
| Maximum length     | 32 code points              |
| Allowed characters | `a-z`, `0-9`, `.`, `_`, `-` |
| Normalisation      | lower-cased                 |

A username is an identifier, not a display name, and the rules decide which
identities can exist and when two inputs name the same one.

**A narrow character set.** It keeps a username safe in URLs and logs, and
rules out homoglyphs and bidirectional control characters. In an auth system,
two usernames that look identical but differ in bytes are a phishing tool.
The cost is that non-Latin scripts cannot be used.

**Lower-casing is canonicalisation, not sanitisation.** `JOAO` and `joao` are
the same identity, stored as `joao`, so no two users differ only by case.

**No trimming.** Removing spaces would silently change what the user typed.
`" alice"` fails the character check instead, which tells the user something
is wrong rather than registering a name they did not type.

**The length bounds** are wide enough not to get in anyone's way, and narrow
enough to keep columns and log lines sensible.

Failures are reported as [validation issues](validation.md), which the API
returns as a `422`.

## Status

`internal/domain/user/status.go`

A value object with a closed set of values, of which there is one, `active`.
`ParseStatus` accepts only known values, so an unexpected value in the
database is an error rather than a state nobody handles.
