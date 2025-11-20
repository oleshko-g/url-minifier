-- +goose Up
-- +goose StatementBegin
ALTER TABLE strings
    ADD CONSTRAINT value_idx
    UNIQUE USING INDEX value_idx;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE strings
    DROP CONSTRAINT IF EXISTS value_idx;
-- +goose StatementEnd
