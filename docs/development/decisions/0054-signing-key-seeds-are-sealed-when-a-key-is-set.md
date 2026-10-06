# Signing key seeds are sealed when an encryption key is set, and plaintext seeds still load

- **Status:** Accepted
- **Date:** 2026-10-06
- **Related:** [0053](0053-signing-keys-are-shared-through-the-database.md), [Signing keys](../../architecture/tokens.md#signing-keys), [`SIGNING_KEY_ENCRYPTION_KEY`](../../configuration.md#signing_key_encryption_key)

## Invariant

When `SIGNING_KEY_ENCRYPTION_KEY` is set, every seed the keyring adds is stored
sealed with AES-256-GCM under it, with the key's generation as additional data,
and the encryption key itself is never stored, logged or echoed. A stored seed
of 32 bytes is plaintext and one of 60 bytes is sealed: a plaintext seed loads
whether or not the key is set, and a sealed one only with the key it was sealed
under.

## Context

[Decision 0053](0053-signing-keys-are-shared-through-the-database.md) put the
signing keys in `signing_keys`, with the 32-byte seed stored as it is, and left
an optional secret to encrypt it for later. Whoever reads the database, or a
backup of it, can therefore sign access tokens that every verifier accepts
while that key is published.

A secret kept outside the database closes that, but v1 is a published
contract: a new variable has to be optional, and the deployments already
running hold plaintext keys that have to keep loading. v2 can require it.

## Decision

**The keyring seals and opens the seeds.** `ed25519.Keyring` already owns
everything about a key's material, so it holds a `seedCipher`: `add` seals the
new seed before `SigningKeyStore.Add`, and `list` opens each stored seed before
restoring the key. `port.SigningKeyStore`, `postgres.SigningKeyRepo` and the
schema do not change: the repository stores whatever bytes it is given.

**AES-256-GCM, from the standard library.** `cipher.NewGCMWithRandomNonce`
over a 32-byte AES key prepends a random 12-byte nonce and appends a 16-byte
tag, so a sealed seed is 60 bytes. A key is added about once a week, far from
the 2^32 messages a random nonce allows. The generation, as 8 bytes big-endian,
is the additional data, so a sealed seed opens only in the row of its
generation.

**The length tells the two forms apart.** 32 bytes is plaintext, 60 is
sealed, and any other length is corrupt. Before this change any seed other
than 32 bytes was already corrupt, so the lengths cannot collide, and no
migration is needed.

**A plaintext seed still loads with the key set.** Once it is set, every key
added is sealed, and the plaintext ones are not rewritten: they are deleted
within about 9 days, as every key is. Enabling the key on a running
deployment therefore needs nothing else, and invalidates no token.

**The variable is optional, and leaving it unset is deprecated.** Unset, seeds
are stored as before, so upgrading changes nothing. v2 requires it. Until then,
startup logs a `WARN` while it is unset, as it does for `RATE_LIMIT=off`, so
the deprecation is seen where operators look.

**The key is 32 bytes in standard base64**, as `openssl rand -base64 32` prints
it. It is parsed with the environment, where an empty value means not set, and
its errors give a byte offset or a size, never the value. The keyring checks
the size again, so a key passed from code fails when the app is built. How to
set it, and how to change it, is in
[Configuration](../../configuration.md#signing_key_encryption_key).

## Alternatives considered

- **Encrypting in `postgres.SigningKeyRepo`**: crypto and a secret inside the
  SQL layer.
- **A `SigningKeyStore` that decorates the repository**: `StoredSigningKey.Seed`
  would carry two meanings, depending on which side of the decorator holds it,
  and there is one more thing to wire.
- **A passphrase through a KDF such as scrypt or Argon2**: invites weak
  secrets, and adds a cost setting.
- **HKDF from a master secret**: nothing else derives a key from it yet.
- **ChaCha20-Poly1305**: no better here, and not in the standard library.
- **16- or 24-byte AES keys too**: one size is simpler to document and to
  check.
- **A `seed_encrypted` column**: a migration, and changes to the model, both
  mappers, the SQL and the test helpers, for what the length already says.
  Adding it to migration 000005 instead would leave a database already at
  version 5 without it, since the version check would pass.
- **A format or version byte**: there is no second format to tell apart.
- **Rejecting plaintext seeds once the key is set**: enabling it would mean
  deleting every stored key, which invalidates outstanding access tokens.
- **Sealing the plaintext rows on load**: a new store method, an `UPDATE` that
  races the other instances, for rows that are gone within days.
- **Documenting the deprecation without the `WARN`**: it would go unseen.
- **Requiring the variable now**: breaks v1.
- **Checking the size in `NewConfig`**: `validation` has only string length
  validators, and its issue codes reach the API's `422` bodies.
- **Rejecting an empty value**: compose files often template an unset variable
  to an empty one, and the `WARN` already flags it.

## Consequences

- Default deployments stay unencrypted until v2, with a `WARN` at every start.
- The key cannot be changed in place. The way out deletes the stored keys,
  which invalidates outstanding access tokens once.
- Every instance needs the same key. One without it, or with another, does not
  start once sealed keys are stored, and one already running fails every sync.
- It protects against reading the database, not writing it: a plaintext row
  written into `signing_keys` still loads. Whoever can write the database can
  already replace a password hash.
- The 32- and 60-byte lengths are part of the stored format, so a second
  format would need another way to tell it apart, such as a column.
