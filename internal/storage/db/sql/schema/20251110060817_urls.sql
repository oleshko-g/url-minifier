-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS string_k_v (
    k TEXT PRIMARY KEY,
    v TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW (),
    deleted_at TIMESTAMPTZ DEFAULT NULL
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS string_k_v;
-- +goose StatementEnd
