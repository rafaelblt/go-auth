# Username rules

- **Status:** Accepted
- **Date:** 2026-09-12
- **Areas:** domain
- **Related:** [0007](0007-new-versus-restore.md), [0013](0013-username-as-separate-entity.md)

## Context

A username here is an identifier, not a display name. The rules decide which
identities can exist and how two inputs are recognised as the same one.

## Decision

3–32 code points, restricted to `a-z`, `0-9`, `.`, `_` and `-`, lower-cased, not
trimmed.

**Restricted character set.** The narrow set keeps a username URL-safe,
log-safe, and free of the ambiguity that homoglyphs and bidirectional control
characters introduce — in an auth system, two usernames that render
identically but differ in bytes are a phishing tool.

**Lower-cased, and that is canonicalisation rather than sanitisation.** Within
the system `JOAO` and `joao` are the same identity, canonically `joao`. Users
should not be able to register names distinguishable only by case.

**No trimming.** Removing leading or trailing whitespace is a strong,
invisible transformation of what the user typed. `" alice"` fails the
allowed-character check instead, which tells the user something is wrong rather
than silently registering a different name than they entered.

**Length bounds** are deliberately generous — wide enough not to annoy anyone,
bounded enough to keep the column and the log lines sane.

## Consequences

Non-Latin scripts cannot be used as usernames.
