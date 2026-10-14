-- CONCURRENTLY keeps logins and refreshes writing while a large table is
-- indexed. It cannot run inside a transaction, so this must stay the only
-- statement in the file. See docs/architecture/persistence.md#migrations.
CREATE INDEX CONCURRENTLY idx_refresh_tokens_session_id_expires_at
    ON refresh_tokens (session_id, expires_at);
