# Password input rules

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain
- **Related:** [0010](0010-username-rules.md), [0012](0012-minimum-password-length.md)

## Context

Whatever the user typed at registration must work at login, and the password
is handed to bcrypt, which silently truncates its input beyond 72 bytes.

## Decision

At least 8 code points, at most 72 bytes, no normalisation, no character
restrictions.

**No allowed-character set.** It is not this service's place to decide which
characters may appear in someone's password. Unlike a username, a password is
never displayed, never part of a URL, and never an identifier — the reasons for
[restricting usernames](0010-username-rules.md) simply do not apply.

**No trimming or Unicode normalisation**, for the same reason: whatever the
user typed must work at login. Normalising here would mean normalising
identically forever, including through any future change of hashing library.

**Maximum 72 bytes**, counted in bytes because that is bcrypt's input limit —
it silently truncates beyond 72, so a longer password would have unused tail
bytes and two different passwords could collide. Capping at exactly the limit
gives users the widest range bcrypt can honour while keeping the rule honest.

This is a considered compatibility bound, not coupling to bcrypt. Replacing the
hasher would not automatically mean changing the constant; 72 bytes is ample
for any password a human will type.

**Minimum 8 code points**, counted in code points because a minimum length is a
statement about how much password there is, not how many bytes it occupies —
an 8-character password of non-Latin characters should not pass a byte-counted
check that an 8-character ASCII one fails.

This record originally set the minimum at 4, a floor on "is this a password at
all" rather than a security recommendation. Why it is 8, and why it is neither
lower nor configurable, is
[Minimum password length](0012-minimum-password-length.md).
