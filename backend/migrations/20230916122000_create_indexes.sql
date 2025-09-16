-- +goose Up
-- +goose StatementBegin
CREATE INDEX idx_links_alias ON links(alias);
CREATE INDEX idx_analytics_link_id ON analytics(link_id);
CREATE INDEX idx_analytics_created_at ON analytics(created_at);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_links_alias;
DROP INDEX IF EXISTS idx_analytics_link_id;
DROP INDEX IF EXISTS idx_analytics_created_at;
-- +goose StatementEnd
