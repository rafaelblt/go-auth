# Design decisions

Choices that are not obvious from reading the code, with the reasoning behind
them. Each decision is its own numbered record, in the style of
[MADR](https://adr.github.io/madr/).

Records are history, not a description of the current system. The documents
under [`architecture/`](../../architecture/overview.md) describe how things
work today and link to the records that explain why. A record stays here when
the code it shaped changes or disappears; its status says whether it still
holds.

## Proposed

Open questions, collected here so they can be reviewed together before 1.0.

| # | Decision | Area |
|---|---|---|
| 0012 | [Minimum password length](0012-minimum-password-length.md) | domain |
| 0013 | [Username as a separate entity](0013-username-as-separate-entity.md) | domain |
| 0029 | [Audit events](0029-audit-events.md) | authentication |
| 0034 | [Protected DTOs](0034-protected-dtos.md) | api |
| 0035 | [Register output shape](0035-register-output-shape.md) | api |
| 0041 | [Database configuration in parts](0041-database-config-in-parts.md) | configuration |

## Index by area

A record that shapes more than one area is listed under each of them.

### Architecture

The service shape and how layers depend on each other. How it works:
[Architecture overview](../../architecture/overview.md).

| # | Decision | Status |
|---|---|---|
| 0001 | [A separate HTTP service rather than a library](0001-separate-http-service.md) | Accepted |
| 0002 | [Ports and adapters](0002-ports-and-adapters.md) | Accepted |
| 0003 | [Split read and write interfaces](0003-split-read-write-interfaces.md) | Accepted |
| 0020 | [Two token types](0020-two-token-types.md) | Accepted |

### Domain

`internal/user`, `internal/password`, `internal/session`, `internal/shared`.
How it works: [Domain model](../../architecture/domain-model.md).

| # | Decision | Status |
|---|---|---|
| 0004 | [`user` and `password` as separate aggregates](0004-user-and-password-separate-aggregates.md) | Accepted |
| 0005 | [`Session` and `RefreshToken` in one package](0005-session-and-refresh-token-one-package.md) | Accepted |
| 0006 | [Typed identities rather than bare UUIDs](0006-typed-identities.md) | Accepted |
| 0007 | [`New` versus `Restore`](0007-new-versus-restore.md) | Accepted |
| 0008 | [`Value()` rather than `String()` for sensitive types](0008-value-rather-than-string.md) | Accepted |
| 0009 | [Defensive copying on reference-typed fields](0009-defensive-copying.md) | Accepted |
| 0010 | [Username rules](0010-username-rules.md) | Accepted |
| 0011 | [Password input rules](0011-password-input-rules.md) | Accepted |
| 0012 | [Minimum password length](0012-minimum-password-length.md) | Proposed |
| 0013 | [Username as a separate entity](0013-username-as-separate-entity.md) | Proposed |
| 0014 | [`Issues` rather than `error` from input constructors](0014-input-constructors-return-issues.md) | Accepted |

### Validation

`internal/validation` and the input constructors that report through it. How
it works: [Validation](../../architecture/domain-model.md#validation).

| # | Decision | Status |
|---|---|---|
| 0009 | [Defensive copying on reference-typed fields](0009-defensive-copying.md) | Accepted |
| 0014 | [`Issues` rather than `error` from input constructors](0014-input-constructors-return-issues.md) | Accepted |
| 0015 | [`Issue` is not an `error`](0015-issue-is-not-an-error.md) | Accepted |
| 0016 | [Validators return `*Issue`](0016-validators-return-issue-pointer.md) | Accepted |
| 0017 | [Separate `MinLength` and `MaxLength`](0017-separate-min-and-max-length.md) | Accepted |
| 0018 | [Explicit length units](0018-explicit-length-units.md) | Accepted |
| 0019 | [`Accumulator.Add` returns nothing](0019-accumulator-add-returns-nothing.md) | Accepted |

### Authentication

Token formats, signing, rotation and reuse detection. How it works:
[Authentication flows](../../architecture/auth-flows.md).

| # | Decision | Status |
|---|---|---|
| 0005 | [`Session` and `RefreshToken` in one package](0005-session-and-refresh-token-one-package.md) | Accepted |
| 0020 | [Two token types](0020-two-token-types.md) | Accepted |
| 0021 | [Ed25519 rather than RSA or HMAC](0021-ed25519-signing.md) | Accepted |
| 0022 | [Minimal JWT claims](0022-minimal-jwt-claims.md) | Accepted |
| 0023 | [Refresh tokens are random bytes, not JWTs](0023-random-refresh-tokens.md) | Accepted |
| 0024 | [Only the refresh token hash is stored](0024-store-refresh-token-hash.md) | Accepted |
| 0025 | [Rotation on every refresh](0025-rotate-on-every-refresh.md) | Accepted |
| 0026 | [Reuse detection revokes the whole session](0026-reuse-revokes-session.md) | Accepted |
| 0027 | [Reuse is checked before expiry](0027-reuse-checked-before-expiry.md) | Accepted |
| 0028 | [Uniform errors on login and refresh](0028-uniform-auth-errors.md) | Accepted |
| 0029 | [Audit events](0029-audit-events.md) | Proposed |

### API

The boundary between `internal/usecase` and `internal/api`. How it works:
[Use cases](../../architecture/overview.md#use-cases) and the
[API reference](../../api/reference.md).

| # | Decision | Status |
|---|---|---|
| 0018 | [Explicit length units](0018-explicit-length-units.md) | Accepted |
| 0028 | [Uniform errors on login and refresh](0028-uniform-auth-errors.md) | Accepted |
| 0030 | [`kind`-based status mapping](0030-kind-based-status-mapping.md) | Accepted |
| 0031 | [The use case adapter](0031-use-case-adapter.md) | Accepted |
| 0032 | [JWKS is a plain handler](0032-jwks-plain-handler.md) | Accepted |
| 0033 | [`use: "sig"` is hard-coded](0033-hardcoded-jwk-use.md) | Accepted |
| 0034 | [Protected DTOs](0034-protected-dtos.md) | Proposed |
| 0035 | [Register output shape](0035-register-output-shape.md) | Proposed |

### Configuration

`internal/config` and the startup checks in `internal/bootstrap`. How it
works: [Configuration](../../configuration.md).

| # | Decision | Status |
|---|---|---|
| 0036 | [Pointer fields for optional config](0036-pointer-fields-for-optional-config.md) | Accepted |
| 0037 | [`LoadConfig` delegates to `NewConfig`](0037-load-config-delegates-to-new-config.md) | Accepted |
| 0038 | [Generic `env[T]`](0038-generic-env-type.md) | Accepted |
| 0039 | [Presets](0039-config-presets.md) | Accepted |
| 0040 | [Startup fails on a schema mismatch](0040-fail-startup-on-schema-mismatch.md) | Accepted |
| 0041 | [Database configuration in parts](0041-database-config-in-parts.md) | Proposed |

### Testing

`internal/testutil` and its sub-packages. How it works:
[Testing](../testing.md).

| # | Decision | Status |
|---|---|---|
| 0042 | [Hand-written fakes](0042-hand-written-fakes.md) | Accepted |
| 0043 | [The test database applies no migrations](0043-test-database-without-migrations.md) | Accepted |
| 0044 | [Wait for the readiness log twice](0044-wait-for-readiness-log-twice.md) | Accepted |

## Writing a record

**When.** A change deserves a record when there was a real alternative, the
choice was deliberate, and the reason cannot be recovered from the code. Most
fixes and refactors apply an existing decision rather than make a new one; the
commit message is enough for those. Limitations and missing features are not
decisions either — they belong in the [roadmap](../roadmap.md).

**How.**

1. Copy [`template.md`](template.md) to `NNNN-short-slug.md`, using the next
   free number. Numbers are never reused, and a file is never renamed.
2. Fill in the metadata, **Context** and **Decision**. Add **Alternatives
   considered** and **Consequences** only when they have something to say; a
   small decision is often just the two required sections.
3. Add the record to the table of every area it belongs to, and to
   [Proposed](#proposed) if it is still open.
4. Link to it from the main documentation, where a reader questioning that
   code would look.

**Areas** are the sections of the index above: `architecture`, `domain`,
`validation`, `authentication`, `api`, `configuration`, `testing`. When a
record fits none of them, add a new section here rather than forcing it into
the nearest one.

## Lifecycle

| Status | Meaning |
|---|---|
| Proposed | An open question. The options are recorded; nothing is chosen. |
| Accepted | In force. |
| Rejected | Proposed, then not adopted. |
| Deprecated | No longer applies, and nothing replaced it — the code it shaped is gone. |
| Superseded by NNNN | Replaced by a later record. |

- **An accepted record is not rewritten.** To change a decision, write a new
  record with **Supersedes** pointing at the old one, then set the old one's
  status to `Superseded by [NNNN](NNNN-slug.md)`. The only other edits allowed
  on an accepted record are the date that goes with a status change and fixes
  to typos or broken links.
- **A proposed record can be edited freely.** To settle it, add a
  **Decision** section, set the status to Accepted (or Rejected), update the
  date, and remove it from [Proposed](#proposed).
- **When a record stops being Accepted**, update its status in the index and
  repoint any links from the main documentation to the record that replaced it.

The **Date** is the day of the last status change.

Records 0001–0044 were split out of the earlier per-area files on 2026-09-12.
Their date is the day of that migration, not of the original decision.
