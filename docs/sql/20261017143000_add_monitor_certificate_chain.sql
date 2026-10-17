-- +goose Up
ALTER TABLE monitor_domain ADD COLUMN ssl_certificate_chain TEXT NOT NULL DEFAULT '[]'; -- 最近采集的SSL证书链JSON
ALTER TABLE monitor_domain ADD COLUMN ssl_certificate_chain_hash VARCHAR(64) NOT NULL DEFAULT ''; -- SSL证书链内容摘要

-- +goose Down
ALTER TABLE monitor_domain DROP COLUMN ssl_certificate_chain_hash;
ALTER TABLE monitor_domain DROP COLUMN ssl_certificate_chain;
