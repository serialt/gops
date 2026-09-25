-- +goose Up
ALTER TABLE host DROP COLUMN ssh_user;

-- +goose Down
ALTER TABLE host ADD COLUMN ssh_user VARCHAR(100) NOT NULL DEFAULT '';
