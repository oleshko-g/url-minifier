-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS strings (
    id TEXT PRIMARY KEY,
    value TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS strings;
-- +goose StatementEnd
