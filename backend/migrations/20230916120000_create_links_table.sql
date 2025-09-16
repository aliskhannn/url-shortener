-- +goose Up
-- +goose StatementBegin
CREATE TABLE links
(
    id         SERIAL PRIMARY KEY,
    url        TEXT               NOT NULL,
    alias      VARCHAR(32) UNIQUE NOT NULL,
    created_at TIMESTAMP          NOT NULL DEFAULT NOW()
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS links;
-- +goose StatementEnd
