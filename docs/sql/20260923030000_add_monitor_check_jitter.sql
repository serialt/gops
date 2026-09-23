-- +goose Up
ALTER TABLE monitor_domain ADD COLUMN check_jitter INTEGER NOT NULL DEFAULT 30;

-- +goose Down
ALTER TABLE monitor_domain DROP COLUMN check_jitter;
