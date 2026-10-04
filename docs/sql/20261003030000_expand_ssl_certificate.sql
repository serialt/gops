-- +goose Up
ALTER TABLE ssl_certificate ADD COLUMN acme_email TEXT NOT NULL DEFAULT ''; /* ACME注册邮箱 */
ALTER TABLE ssl_certificate ADD COLUMN key_algorithm TEXT NOT NULL DEFAULT 'rsa2048'; /* 密钥算法 */
ALTER TABLE ssl_certificate ADD COLUMN cloud_account_id INTEGER; /* 云账号ID */
ALTER TABLE ssl_certificate ADD COLUMN cloud_cert_id TEXT NOT NULL DEFAULT ''; /* 云厂商证书ID */
ALTER TABLE ssl_certificate ADD COLUMN cert_chain TEXT NOT NULL DEFAULT ''; /* 证书链 */
ALTER TABLE ssl_certificate ADD COLUMN not_before DATETIME; /* 生效时间 */
ALTER TABLE ssl_certificate ADD COLUMN fingerprint TEXT NOT NULL DEFAULT ''; /* 证书指纹 */
ALTER TABLE ssl_certificate ADD COLUMN dns_provider_id INTEGER; /* DNS服务商配置ID */
ALTER TABLE ssl_certificate ADD COLUMN last_renew_at DATETIME; /* 最近续期时间 */

ALTER TABLE ssl_dns_provider ADD COLUMN phone TEXT NOT NULL DEFAULT ''; /* 联系电话 */

CREATE TABLE IF NOT EXISTS ssl_deploy_config (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 部署配置ID */
  certificate_id INTEGER NOT NULL REFERENCES ssl_certificate(id) ON DELETE CASCADE, /* 证书ID */
  name TEXT NOT NULL, /* 配置名称 */
  deploy_type TEXT NOT NULL, /* 部署类型 */
  target_config TEXT NOT NULL DEFAULT '{}', /* 目标配置JSON */
  auto_deploy BOOLEAN NOT NULL DEFAULT TRUE, /* 续期后自动部署 */
  enabled BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  last_deploy_at DATETIME, /* 最近部署时间 */
  last_deploy_ok BOOLEAN NOT NULL DEFAULT FALSE, /* 最近部署是否成功 */
  last_error TEXT NOT NULL DEFAULT '', /* 最近错误 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_ssl_deploy_config_certificate ON ssl_deploy_config(certificate_id);

CREATE TABLE IF NOT EXISTS ssl_task (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* SSL任务ID */
  certificate_id INTEGER NOT NULL REFERENCES ssl_certificate(id) ON DELETE CASCADE, /* 证书ID */
  task_type TEXT NOT NULL, /* 任务类型 */
  status TEXT NOT NULL DEFAULT 'pending', /* 任务状态 */
  trigger_type TEXT NOT NULL DEFAULT 'manual', /* 触发类型 */
  started_at DATETIME, /* 开始时间 */
  finished_at DATETIME, /* 完成时间 */
  error_message TEXT NOT NULL DEFAULT '', /* 错误信息 */
  result TEXT NOT NULL DEFAULT '', /* 任务结果JSON */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_ssl_task_certificate ON ssl_task(certificate_id);
CREATE INDEX IF NOT EXISTS idx_ssl_task_status ON ssl_task(status);

-- +goose Down
DROP TABLE IF EXISTS ssl_task;
DROP TABLE IF EXISTS ssl_deploy_config;
ALTER TABLE ssl_dns_provider DROP COLUMN phone;
ALTER TABLE ssl_certificate DROP COLUMN last_renew_at;
ALTER TABLE ssl_certificate DROP COLUMN dns_provider_id;
ALTER TABLE ssl_certificate DROP COLUMN fingerprint;
ALTER TABLE ssl_certificate DROP COLUMN not_before;
ALTER TABLE ssl_certificate DROP COLUMN cert_chain;
ALTER TABLE ssl_certificate DROP COLUMN cloud_cert_id;
ALTER TABLE ssl_certificate DROP COLUMN cloud_account_id;
ALTER TABLE ssl_certificate DROP COLUMN key_algorithm;
ALTER TABLE ssl_certificate DROP COLUMN acme_email;
