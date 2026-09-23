# Decision records

Decisions whose reasoning cannot be read from the code, and is too detailed
for the main documentation. Each decision is a numbered record, in the style
of [MADR](https://adr.github.io/madr/).

The main documentation says how the system works and, briefly, why. A record
goes further, for one decision: the forces behind it, the alternatives
weighed, and what it costs.

## Records

| # | Decision | Areas |
|---|---|---|
| 0034 | [DTOs are protected structs](0034-protected-dtos.md) | api |
| 0045 | [Login verifies a dummy hash when the account is missing](0045-login-dummy-hash.md) | authentication |
| 0046 | [Refresh token use is guarded at write](0046-refresh-token-use-guarded-at-write.md) | authentication, architecture |
| 0047 | [A refresh that loses the race for its token is reuse](0047-lost-refresh-race-is-reuse.md) | authentication |
| 0048 | [A duplicate username is reported by the user writer](0048-duplicate-username-reported-by-the-writer.md) | architecture, domain |
| 0049 | [The refresh token secret belongs to the domain](0049-refresh-token-secret-in-the-domain.md) | architecture, domain, authentication |

The numbers have gaps. Records 0001 to 0044 were short rationales, and on
2026-09-19 they were folded into the main documentation, next to what they
explain. A reference to one of them, in an older commit message, can be
followed in the git history.

## When to write one

A decision gets a record when all of these hold:

- there was a real alternative, and the choice was deliberate;
- the reasoning cannot be recovered from the code;
- it does not fit in a code comment, because it spans several packages or has
  to weigh alternatives;
- the main documentation could give it a sentence or two, not the whole
  argument.

Otherwise it goes somewhere else:

- a reason that takes a sentence or two goes in the main documentation, next
  to what it explains;
- a reason tied to one function goes in a comment on that function;
- a limitation goes in [Limitations](../../limitations.md);
- an idea or an open question goes in
  [GitHub Issues](https://github.com/rafaelblt/go-auth/issues). A record
  describes a decision that was made.

Most fixes and refactors apply an existing decision rather than make a new
one, and their commit message is enough. [Commits](../commits.md) says what
such a message should hold.

## Writing one

A record is written together with the change it describes, in the same
commit.

1. Copy [`template.md`](template.md) to `NNNN-short-slug.md`, with the next
   free number. Numbers are never reused, and a file is never renamed.
2. Fill in the metadata, **Context** and **Decision**. Add **Alternatives
   considered** and **Consequences** only when they have something to say.
3. Add the record to [Records](#records).
4. Link to it from where a reader questioning that code would look: the
   section of the main documentation, and a comment on the code it shaped.

## Lifecycle

| Status | Meaning |
|---|---|
| Accepted | In force. |
| Deprecated | No longer applies, and nothing replaced it: the code it shaped is gone. |
| Superseded by NNNN | Replaced by a later record. |

- **An accepted record is not rewritten.** To change a decision, write a new
  record with **Supersedes** pointing at the old one, and set the old one's
  status to `Superseded by [NNNN](NNNN-slug.md)`. The only other edits allowed
  on an accepted record are the date that goes with a status change, and fixes
  to typos or broken links.
- **When a record stops being Accepted**, update its status in
  [Records](#records), and repoint the links to it from the main documentation
  and the code.

The **Date** is the day of the last status change.
