# Decision records

A record explains one decision whose reasoning does not fit next to the code it
shaped: it spans packages, and there were real alternatives to weigh. Each
opens with its **Invariant**, what a change must not break for the decision to
hold, and a comment on the code it shaped links to it, because that code often
looks removable.

| # | Decision |
|---|---|
| 0034 | [A mapper is the only way to build a DTO](0034-protected-dtos.md) |
| 0045 | [A missing account costs a bcrypt comparison too](0045-login-dummy-hash.md) |
| 0050 | [Refresh token use is settled at write, and losing that race is reuse](0050-refresh-token-use-is-settled-at-write.md) |
| 0053 | [Signing keys are shared through the database, and the clock decides which one signs](0053-signing-keys-are-shared-through-the-database.md) |
| 0055 | [Login locks the password it checked until its session is saved](0055-login-locks-the-password-it-checked.md) |

The numbers have gaps. A record whose reasoning fits in a few sentences was
folded into the document of the code it explains, and a replaced one was
deleted; the git history keeps both, and an old commit message's "See decision
NNNN" can be followed there.

## When to write one

When the reason for a decision takes more than a few sentences, spans packages,
and weighs alternatives that were really considered. Otherwise it goes in the
document of the package, next to what it explains, or in a comment on the
function, and most changes apply an existing decision rather than make one.

## Keeping them

A record has a title that states the decision as a fact, a
`**Status:** Accepted` line, the **Invariant**, the **Context** and the
**Decision**, and, when they have something to say, the **Alternatives** that
were really weighed and the **Consequences**. It states what the main documents
already explain by linking to them.

A record describes the decision in force, so it is kept true like any other
document: when the decision changes, the record is edited, and when the code it
shaped is gone, the record is deleted and the links to it repointed.
`tests/docs` fails when an accepted record is mentioned by no `.go` file.
