-- +goose Up
-- +goose StatementBegin
CREATE TABLE analytics
(
    id          BIGSERIAL PRIMARY KEY,
    link_id     INT       NOT NULL REFERENCES links (id) ON DELETE CASCADE,
    user_agent  TEXT,
    device_type VARCHAR(32),
    os          VARCHAR(64),
    browser     VARCHAR(64),
    ip_address  INET,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS analytics;
-- +goose StatementEnd
