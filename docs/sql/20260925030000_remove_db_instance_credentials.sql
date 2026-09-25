-- +goose Up
ALTER TABLE db_instance DROP COLUMN username;
ALTER TABLE db_instance DROP COLUMN password;

-- +goose Down
ALTER TABLE db_instance ADD COLUMN username VARCHAR(255) NOT NULL DEFAULT '';
ALTER TABLE db_instance ADD COLUMN password TEXT NOT NULL DEFAULT '';
