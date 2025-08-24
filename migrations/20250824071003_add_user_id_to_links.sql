-- +goose Up
ALTER TABLE links
ADD COLUMN user_id BIGINT REFERENCES users(id) ON DELETE CASCADE;

-- +goose Down
ALTER TABLE links
DROP COLUMN IF EXISTS user_id;