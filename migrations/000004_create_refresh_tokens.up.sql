CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL
        REFERENCES sessions(id)
        ON DELETE CASCADE,
    parent_id UUID UNIQUE NULL
        REFERENCES refresh_tokens(id)
        ON DELETE SET NULL,
    hash BYTEA UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at TIMESTAMPTZ NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
