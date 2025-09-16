-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE analytics
(
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    link_id     UUID NOT NULL REFERENCES links(id) ON DELETE CASCADE,
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
DROP EXTENSION IF EXISTS "uuid-ossp";
-- +goose StatementEnd
