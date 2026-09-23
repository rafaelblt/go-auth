# The refresh token secret belongs to the domain

- **Status:** Accepted
- **Date:** 2026-09-18
- **Areas:** architecture, domain, authentication
- **Related:** [Token model](../../architecture/tokens.md#token-model), [Code conventions](../../architecture/conventions.md#value-objects)

## Context

A refresh token's format (32 bytes from `crypto/rand`, base64url) and its
storage (only `SHA-256(token)`) are fixed by the
[token model](../../architecture/tokens.md#token-model). Both lived in `internal/infra/refreshtoken`, behind two ports:
`RefreshTokenGenerator`, which produced the raw token and its hash, and
`RefreshTokenResolver`, which decoded a raw token, hashed it and called
`RefreshTokenReader.FindByHash`.

That package imported only the standard library and did no I/O of its own.
It had no second implementation, and no plausible one: another length or
encoding would be a change to the token format, not a different adapter. The ports
bought nothing a port exists for.

The split also left the domain incomplete. `RefreshTokenHash` was a domain
value object, but nothing in the domain knew what it was the hash of: it
accepted any non-empty bytes, `NewRefreshToken` took whatever hash the
caller passed, and the raw token was a bare `string`. The rule tying the
token to its hash existed only in infra.

## Decision

`session.RefreshTokenSecret` is the credential handed to the client.

**Creation goes through the entity.** `NewRefreshToken` generates the secret
and returns it alongside the token:

```go
func NewRefreshToken(params RefreshTokenCreationParams) (*RefreshToken, RefreshTokenSecret, error)
```

`RefreshTokenCreationParams` has no hash field, and the secret's constructor
is unexported, so a new token cannot carry a hash that does not come from a
secret, and callers do not assemble the token from parts. The token keeps
only the hash; the returned secret is the one chance to hand it out. It is a
return value rather than a field so it never travels into the writer, a log
line or a restored token.

**Resolution parses, then looks up.** `ParseRefreshTokenSecret` turns the
client's string into a secret, checking the encoding and the length, and
`Hash()` gives the lookup key. Like the other sensitive types, the secret
exposes `Value()` rather than `String()`
([conventions](../../architecture/conventions.md#value-rather-than-string)), so it never prints itself
through a `%v`.

**A zero secret panics.** `Value()` and `Hash()` panic on a zero secret, unlike
the other value objects, which return their zero quietly. A zero secret cannot
come from input: `Parse` rejects what is malformed, and `NewRefreshToken`
returns one only alongside an error. It exists only when an error was ignored
or the struct was built by hand. The other value objects' zeros are caught
further on, by `Restore` or by the DTO mappers; a zero hash is not. It would
reach `FindByHash` as a token that is simply not found, and a programming
error would be answered as `ErrTokenInvalid`, with no log and nothing in the
tests to tell it from the correct behaviour. The use case calls
`RefreshTokenReader.FindByHash` itself: the lookup is I/O, which the domain
does not do, and it was already a port. The name is `Parse`, not `Restore`:
the secret is never stored, so there is nothing to restore
([conventions](../../architecture/conventions.md#two-constructors)), and `Parse` is what the other string →
value object constructors in the domain are called.

**The hash is exactly a SHA-256 digest.** `NewRefreshTokenHash` requires
`sha256.Size` bytes. This is a structural invariant, so `Restore` applies it
too; every stored hash already satisfies it.

`RefreshTokenGenerator`, `RefreshTokenResolver`, `RefreshTokenGenerated` and
`internal/infra/refreshtoken` are removed.

## Alternatives considered

- **Keeping the ports** — they abstracted policy, not technology, and forced
  tests to fake two layers for one lookup.
- **An exported `NewRefreshTokenSecret`, passed to `NewRefreshToken`** — the
  caller would assemble the token from parts, and nothing would stop it
  passing an unrelated hash.
- **Keeping the raw token inside the hash or the entity, readable once** —
  it mixes the credential with the digest that storing only the hash separates, a
  "read once" flag does not survive copying a value type, and the secret
  would reach the writer and any log of the entity.

## Consequences

- A malformed refresh token is rejected as `ErrTokenInvalid` (`401`). Before,
  a string that did not decode as base64url was an unexpected error (`500`).
- Tests can no longer choose the secret a use case generates. They check the
  real invariant instead: the returned secret hashes to the stored hash.
- Generating a secret still returns an error, although `crypto/rand.Read`
  cannot fail since Go 1.24. `NewRefreshToken` returns an error anyway, so
  keeping it costs callers no extra branch.
