-- +goose Up
ALTER TABLE monitor_domain ADD COLUMN request_client VARCHAR(20) NOT NULL DEFAULT 'chrome';

-- +goose Down
ALTER TABLE monitor_domain DROP COLUMN request_client;
