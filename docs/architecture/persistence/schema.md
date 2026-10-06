# Schema

Five tables: one per entity, plus `signing_keys`, which holds the keys that
sign access tokens and stands alone, with no foreign key to or from any other.

```
users ──1:0..1── passwords
  │
  └──1:N── sessions ──1:N── refresh_tokens
                                  │
                                  └── parent_id ──► refresh_tokens
```

The entities behind them are in [Domain](../domain/README.md).

## `users`

```sql
CREATE TABLE users (
    id         UUID PRIMARY KEY,
    username   TEXT UNIQUE NOT NULL,
    status     TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

`username` is unique, and that constraint settles two concurrent
registrations of the same name. The insert targets it with
`ON CONFLICT (username) DO NOTHING`, so the loser inserts nothing and is told
the username is taken
([decision 0048](../../development/decisions/0048-duplicate-username-reported-by-the-writer.md)).
Values are always lower case: the domain canonicalises them first.

## `passwords`

```sql
CREATE TABLE passwords (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    hash       TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

`user_id` is `UNIQUE`, which is what enforces one password per user: the
domain does not. Deleting a user deletes its password.

`hash` holds the whole bcrypt string, with its algorithm, cost and salt, so
each password records its own cost, and changing `BCRYPT_COST` does not break
existing logins. Nothing rehashes a stored password, so old passwords keep the
old cost.

## `sessions`

```sql
CREATE TABLE sessions (
    id         UUID PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    revoked_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
```

`revoked_at` is `NULL` while the session is active. Revocation is a timestamp
rather than a boolean, so *when* it happened is recorded.

## `refresh_tokens`

```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY,
    session_id UUID NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    parent_id  UUID UNIQUE NULL REFERENCES refresh_tokens(id) ON DELETE SET NULL,
    hash       BYTEA UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
```

- **`hash`** is the raw SHA-256 digest, never the token. Tokens are looked up
  by it, so it needs the index its `UNIQUE` creates, and two equal digests
  would be a serious bug.
- **`parent_id`** links each token to the one it replaced, and is `NULL` for
  the token created at login. It is `UNIQUE`, so a token has at most one
  successor, and the rotation chain can never fork.
- **`used_at`** is `NULL` until the token is spent. `RefreshTokenRepo.MarkUsed`
  sets it with `WHERE used_at IS NULL`, so of two concurrent uses only one is
  applied
  ([decision 0050](../../development/decisions/0050-refresh-token-use-is-settled-at-write.md)).

Nothing deletes rows from this table; see
[Limitations](../../limitations.md#refresh-token-rows-are-never-deleted).

## `signing_keys`

```sql
CREATE TABLE signing_keys (
    generation BIGINT PRIMARY KEY,
    seed       BYTEA NOT NULL,
    active_at  TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
```

- **`generation`** orders the keys: 1, 2, and so on. As the primary key, it
  settles two instances adding the same next key: the insert targets it with
  `ON CONFLICT (generation) DO NOTHING`, so the second inserts nothing.
- **`seed`** is the key's Ed25519 private key, either as its 32 bytes,
  [unencrypted](../../limitations.md#signing-keys-are-stored-in-the-database),
  or, when [`SIGNING_KEY_ENCRYPTION_KEY`](../../configuration.md#signing_key_encryption_key)
  is set, sealed with AES-256-GCM into 60 bytes: a 12-byte nonce, the sealed
  seed and a 16-byte tag, with the generation as additional data. The length
  tells the two apart.
- **`active_at`** is when the key starts signing.
- **`created_at`** is when it was added, and from then on published.

The key ID (`kid`) is not stored: it is derived from the key. A row is
deleted once the next key has been signing for 25 hours; the whole lifecycle
is in [Signing keys](../tokens.md#signing-keys).
