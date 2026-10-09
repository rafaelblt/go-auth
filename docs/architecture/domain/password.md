# Password

`internal/domain/password`

## Password

`internal/domain/password/password.go`

An entity separate from `User`, holding a hash and a reference to its user.

| Field                     | Type        |
| ------------------------- | ----------- |
| `id`                      | `ID`        |
| `userID`                  | `user.ID`   |
| `hash`                    | `Hashed`    |
| `createdAt` / `updatedAt` | `time.Time` |

As its own entity, it says that a user *may* have a password, not that it
must. Another kind of credential can be added later as another entity. A
generic `Credential` entity, meant to cover every kind, was tried first and
dropped: its fields could only describe credentials that did not exist yet.

Nothing in the domain enforces one password per user. The `UNIQUE` constraint
on `passwords.user_id` does.

`ChangeHash(hash, at)` replaces the hash and sets `updatedAt`, and refuses a
zero hash. The ID and `createdAt` stay: a new password is the same row with
another hash, so the user still has one. `PasswordWriter.UpdateHash` persists
the change only while the stored hash is still the one it replaced, and
returns `ErrHashChanged` otherwise, so of two concurrent changes only one
applies ([guarded writes](../persistence/repositories.md#guarded-writes)).

## Plain

`internal/domain/password/plain.go`

A password as the user typed it.

| Rule               | Value         |
| ------------------ | ------------- |
| Minimum length     | 8 code points |
| Maximum length     | 72 bytes      |
| Normalisation      | none          |
| Allowed characters | any           |

Whatever the user typed at registration must work at login, so nothing is
changed and nothing is forbidden.

**Any character.** The reasons for restricting usernames do not apply: a
password is never displayed, never part of a URL, and never identifies
anyone. Which characters appear in it is not this service's business.

**No trimming or Unicode normalisation.** Normalising would mean normalising
the same way forever, through any future change of hashing library.

**At most 72 bytes**, counted in bytes because that is bcrypt's input limit.
Bcrypt silently ignores anything beyond 72 bytes, so two longer passwords that
share their first 72 bytes would match each other. Capping at the limit gives
users the most bcrypt can honour. 72 bytes is also more than any password a
person types, so replacing bcrypt would not need to change it.

**At least 8 code points**, counted in code points because a minimum is about
how much password there is, not how many bytes it takes: counted in bytes, 4
characters of a non-Latin script would pass where 4 ASCII characters fail. Eight is
what NIST SP 800-63B requires of a password the user chooses, and it is the
default nearly every deployment gets, so a lower one would be a liability
by default. A shorter secret, like a PIN, is only safe with what makes its
few possibilities hard to try: attempt limits per account, lockout, rate
limiting. This service has none of them.

**Not configurable.** The validators are built once from constants, so a
runtime setting would have to be threaded through `NewPlain` and every
caller, or kept in package state, for a value almost nobody changes. A
deployment that needs another minimum changes `PlainMinCodePoints` in its
copy of the code.

This is a minimum, not a strength policy: there are no composition rules
(a digit, a symbol), and a minimum length is the rule that remains when those
are dropped.

`Plain` exposes `Value()`, not `String()`, so it cannot end up in a log line
by accident.

## Hashed

`internal/domain/password/hashed.go`

The hasher's output. The only rule is that it is not empty: a trusted
component generates it, so there is nothing else to check. It exposes
`Value()`, not `String()`.
