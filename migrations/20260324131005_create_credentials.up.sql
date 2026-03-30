CREATE TABLE credentials (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL
        REFERENCES users(id)
        ON DELETE CASCADE,
    kind VARCHAR(10) NOT NULL,
    provider VARCHAR(10) NOT NULL,
    secret TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
