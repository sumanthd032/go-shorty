-- File: migrations/YYYYMMDDHHMMSS_create_links_table.sql

-- +goose Up
-- +goose StatementBegin
CREATE TABLE links (
    id BIGSERIAL PRIMARY KEY,
    alias TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL,
    password_hash BYTEA,
    expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd