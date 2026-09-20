-- +goose Up
ALTER TABLE monitor_domain
ADD COLUMN ssl_root_ca VARCHAR(255) NOT NULL DEFAULT ''; -- SSL 证书根 CA 名称

-- +goose Down
ALTER TABLE monitor_domain DROP COLUMN ssl_root_ca;
