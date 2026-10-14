# Domain

`internal/domain/user`, `internal/domain/password`, `internal/domain/session`,
`internal/validation`, `internal/shared`

Four entities in three packages, the validation types they share, and a few
generic helpers. Every type follows the [code conventions](conventions.md):
fields are unexported, and state changes only through methods that enforce the
type's own rules.

## Entities

```
User ──1:0..1── Password
 │
 └──1:N── Session ──1:N── RefreshToken
                              │
                              └── parent_id ──► RefreshToken  (rotation chain)
```

Every identity (`user.ID`, `password.ID`, `session.SessionID`,
`session.RefreshTokenID`) wraps `shared.EntityID`, a UUID, version 4 when
generated, that rejects the nil UUID. Functions often take several IDs at once,
and with a bare `uuid.UUID` swapping two of them would compile and fail later as
"not found"; wrapped, it does not compile.

`user` and `password` are separate packages because a user *may* have a
password: another kind of credential can be added later as its own entity,
without changing `User`. `session` holds both `Session` and `RefreshToken`
because they change together: revoking a session invalidates its tokens, and
reuse detection reads a token and revokes its session in one operation. In
separate packages, each would import the other.

## `user`

`User` is the account: created by registration, found by username at login.
Its `Status` is a closed set with one value, `active`; `ParseStatus` rejects
anything else, so an unexpected value in the database is an error rather than
a state nobody handles.

`Username` is an identifier, not a display name:

| Rule | Value |
|---|---|
| Length | 3 to 32 code points |
| Characters | `a-z`, `0-9`, `.`, `_`, `-` |
| Normalisation | lower-cased |

- **A narrow character set** keeps a username safe in URLs and logs, and rules
  out homoglyphs and bidirectional control characters: two usernames that look
  the same but differ in bytes are a phishing tool. The cost is that non-Latin
  scripts cannot be used.
- **Lower-casing is canonicalisation.** `JOAO` and `joao` are the same identity,
  stored as `joao`.
- **No trimming.** `" alice"` fails the character check rather than registering
  a name the user did not type.

## `password`

`Password` is an entity of its own, holding a hash and its user's ID. One
password per user is enforced by the `UNIQUE` on `passwords.user_id`, not by the
domain. `ChangeHash` replaces the hash and keeps the ID and `createdAt`: a new
password is the same row with another hash, and the write that persists it is
[guarded](persistence.md#guarded-writes).

`Plain` is a password as the user typed it:

| Rule | Value |
|---|---|
| Minimum length | 8 code points |
| Maximum length | 72 bytes |
| Characters | any |
| Normalisation | none |

Whatever was typed at registration must work at login, so nothing is changed
and nothing is forbidden:

- **Any character.** A password is never displayed, never part of a URL, and
  never identifies anyone.
- **No trimming or Unicode normalisation.** Normalising would mean normalising
  the same way forever, through any change of hashing library.
- **At most 72 bytes**, bcrypt's input limit. Bcrypt ignores anything past it,
  so two longer passwords sharing their first 72 bytes would match each other.
- **At least 8 code points**, the minimum chosen for this service. A minimum
  counts code points, because it is about how much password there is: counted
  in bytes, 4 characters of a non-Latin script would pass where 4 ASCII
  characters fail.
- **Not configurable.** The validators are built once from constants, and
  `PlainMinCodePoints` is the one to change in a fork that needs another
  minimum. Raising it locks out the users whose passwords are shorter
  ([Compatibility](../../CONTRIBUTING.md#compatibility)).

There are no composition rules (a digit, a symbol): a minimum length is the
whole policy.

`NewPlainReplacing(value, current)` applies the same rules and then refuses
`current` itself, with `same_as_current`: changing a password to the same one
would change nothing but sign the user out everywhere. It compares the two byte
for byte, and only against the current password: no history is kept.

`Hashed` is the hasher's output. A trusted component generates it, so the only
rule is that it is not empty.

## `session`

**`Session`** is a login, revoked by reuse detection or by a password change.
`Revoke` is idempotent in memory, but `SessionWriter.Update` is not guarded:
two concurrent requests that each read an active session and revoke it both
write, and the later write wins. Either way the session is revoked, so this is
harmless. `SessionWriter.RevokeAllByUserID` revokes in SQL instead, so it
reaches sessions nobody loaded, and sets `revoked_at` only where it is still
`NULL`, keeping an earlier revocation's time.

**`RefreshToken`** is one link in a rotation chain: login creates one without a
parent, and each refresh creates one whose `parentID` is the token it replaced.
`NewRefreshToken` generates the token's secret and returns it alongside the
token, which keeps only the secret's hash. Its params carry no hash and the
secret's constructor is unexported, so a stored hash always comes from a real
secret, and the secret, being a return value rather than a field, never reaches
a writer, a log line or a restored token.

`Use` spends the token, and checks "already used" before "expired", so
replaying a token that was spent and has since expired still counts as reuse
and revokes the session. The other way round, someone holding an old stolen
token could wait for it to expire, replay it, and never trip the alarm. `Use`
decides in memory, on a token read before the write, so the write checks again
([decision 0050](../development/decisions/0050-refresh-token-use-is-settled-at-write.md)).

**`RefreshTokenSecret`** is what the client holds: 32 bytes from `crypto/rand`,
which `Value()` encodes in base64url. `ParseRefreshTokenSecret` reads a
client's string back, rejecting a bad encoding or length, and `Hash()` gives
the `RefreshTokenHash` to look the token up by. It is `Parse`, not `Restore`,
because the secret is never stored. `Value()` and `Hash()` panic on a zero
secret: one can only come from an ignored error, and its hash would otherwise
be looked up as a token that simply does not exist, hiding the bug.

**`RefreshTokenHash`** is the SHA-256 of a secret, exactly 32 bytes, and the
only form of the token that is stored. **`AccessToken`** is a signed JWT, not
empty, typed so it cannot be mixed up with another string on its way out. What
the two tokens are for: [Tokens](tokens.md).

## `validation`

The types for input rules, with two callers: the domain, whose failures the
API answers with a `422`, and `internal/config`, whose failures become the
[startup error](../configuration.md#startup-validation). It is not a domain
package, and a change to it changes both reports.

An **`Issue`** is one broken rule: a `code` and a `details` map, such as
`too_short` with `{"min": 3, "unit": "code_point"}`. It does not implement
`error`: nothing treats one issue as an error, and since `details` is a map, an
`Issue` is not comparable, so `errors.Is` against one would quietly return
`false`.

A **`Validator[T]`** is `func(T) *Issue`, where `nil` means the value passed.
`Validate(value, validators...)` runs every one and collects every failure.

| Validator | Issue | `details` | Used by |
|---|---|---|---|
| `MinLength(n, unit)` | `too_short` | `min`, `unit` | domain |
| `MaxLength(n, unit)` | `too_long` | `max`, `unit` | domain |
| `AllowedChars(set)` | `invalid_characters` | — | domain |
| `Required[T]()` | `required` | — | config |
| `Positive[T]()` | `not_positive` | — | config |

`internal/config` also writes its own validators, `allowedLogFormat` and
`allowedRateLimitLevel`, which return `not_allowed` with the accepted values in
`allowed`. `same_as_current` has no validator, since it compares two values:
`password.NewPlainReplacing` returns it itself.

`MinLength` and `MaxLength` are separate, so each produces one code; a single
`Length(min, max)` would take two adjacent `int`s and invite swapping them.
Every length is counted in an explicit unit, `code_point` or `byte`, which
reaches the API response: Go's `len` counts bytes, JavaScript's `.length`
UTF-16 code units, and a user characters, so with the unit implicit, client and
server would disagree.

The **`Accumulator`** collects the issues of several fields. `Add` returns
nothing, so there is no `if err != nil` after each field, and `Err()` returns a
`ValidationError` holding every `FieldError`, or `nil`. This is how registration
reports a bad username and a bad password in one response, and how a startup
failure lists every misconfigured variable.

## `shared`

Helpers with no domain meaning: the `EntityID` every identity wraps, a `Set`
(the allowed username characters), and pointer helpers for optional fields.
