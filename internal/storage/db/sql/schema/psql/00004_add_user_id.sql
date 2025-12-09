-- +goose Up
ALTER TABLE strings
ADD COLUMN user_id TEXT;


-- +goose Down
ALTER TABLE strings
DROP COLUMN user_id;
