-- +goose Up
ALTER TABLE monitor_domain
ADD COLUMN verify_certificate BOOLEAN NOT NULL DEFAULT TRUE; -- HTTPS 是否校验证书链和域名

-- +goose Down
ALTER TABLE monitor_domain DROP COLUMN verify_certificate;
