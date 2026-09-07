CREATE TABLE password_credentials (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE
        REFERENCES users(id)
        ON DELETE CASCADE,
    hash TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
