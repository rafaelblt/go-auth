CREATE TABLE signing_keys (
    generation BIGINT PRIMARY KEY,
    seed BYTEA NOT NULL,
    active_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL
);
