-- +goose Up
-- +goose StatementBegin
CREATE TABLE clicks (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    clicked_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ip_address TEXT,
    user_agent TEXT,
    referrer TEXT -- This line was likely missing before
);

CREATE INDEX idx_clicks_link_id ON clicks(link_id);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS clicks;
-- +goose StatementEnd