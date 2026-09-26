-- +goose Up
-- Gops SQLite 最终基线结构。历史建表、字段调整与数据修复迁移已合并至此文件。

CREATE TABLE IF NOT EXISTS role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  code VARCHAR(64) NOT NULL UNIQUE,
  name VARCHAR(64) NOT NULL,
  description VARCHAR(255),
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_code_active ON role(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username VARCHAR(64) NOT NULL UNIQUE,
  password_hash VARCHAR(255) NOT NULL,
  nickname VARCHAR(64) NOT NULL,
  email VARCHAR(255) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username_active ON user(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_active ON user(email) WHERE deleted_at IS NULL AND email <> '';

CREATE TABLE IF NOT EXISTS user_role (
  user_id INTEGER NOT NULL,
  role_id INTEGER NOT NULL,
  PRIMARY KEY (user_id, role_id)
);
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON user_role(role_id);

CREATE TABLE IF NOT EXISTS organization (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  parent_id INTEGER REFERENCES organization(id) ON DELETE RESTRICT,
  name VARCHAR(64) NOT NULL,
  code VARCHAR(64) NOT NULL,
  manager VARCHAR(64) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME,
  sort_order INTEGER NOT NULL DEFAULT 0
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_code_active ON organization(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organization_parent_id ON organization(parent_id);

CREATE TABLE IF NOT EXISTS organization_user (
  organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  PRIMARY KEY (organization_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_organization_user_user_id ON organization_user(user_id);

INSERT INTO organization(name, code, manager, status)
SELECT '默认组织', 'root', '', TRUE
WHERE NOT EXISTS (SELECT 1 FROM organization WHERE code = 'root' AND deleted_at IS NULL);

INSERT INTO organization_user(organization_id, user_id)
SELECT root.id, users.id
FROM organization AS root
CROSS JOIN user AS users
WHERE root.code = 'root' AND root.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM organization_user AS relation
    WHERE relation.organization_id = root.id AND relation.user_id = users.id
  );

CREATE TABLE IF NOT EXISTS system_config (
  config_key VARCHAR(128) PRIMARY KEY,
  config_value TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
INSERT INTO system_config(config_key, config_value) VALUES ('system_name', 'Gops') ON CONFLICT(config_key) DO NOTHING;
INSERT INTO system_config(config_key, config_value) VALUES ('system_description', '') ON CONFLICT(config_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS asset_group (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  code VARCHAR(50) NOT NULL,
  parent_id BIGINT NOT NULL DEFAULT 0,
  description VARCHAR(500) NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_group_code_active ON asset_group(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS credential (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  credential_type VARCHAR(20) NOT NULL,
  username VARCHAR(100) NOT NULL DEFAULT '',
  password TEXT NOT NULL DEFAULT '',
  private_key TEXT NOT NULL DEFAULT '',
  passphrase TEXT NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS cloud_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  provider VARCHAR(50) NOT NULL,
  access_key TEXT NOT NULL DEFAULT '',
  secret_key TEXT NOT NULL DEFAULT '',
  region VARCHAR(100) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS host (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  group_id BIGINT REFERENCES asset_group(id) ON DELETE SET NULL,
  host_type VARCHAR(20) NOT NULL DEFAULT 'self',
  ip VARCHAR(255) NOT NULL,
  port INTEGER NOT NULL DEFAULT 22,
  credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL,
  cloud_account_id BIGINT REFERENCES cloud_account(id) ON DELETE SET NULL,
  tags VARCHAR(500) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT -1,
  last_seen_at DATETIME,
  os VARCHAR(100) NOT NULL DEFAULT '',
  kernel VARCHAR(100) NOT NULL DEFAULT '',
  arch VARCHAR(50) NOT NULL DEFAULT '',
  cpu_info TEXT NOT NULL DEFAULT '',
  cpu_cores INTEGER NOT NULL DEFAULT 0,
  cpu_usage REAL NOT NULL DEFAULT 0,
  memory_total BIGINT NOT NULL DEFAULT 0,
  memory_used BIGINT NOT NULL DEFAULT 0,
  memory_usage REAL NOT NULL DEFAULT 0,
  disk_total BIGINT NOT NULL DEFAULT 0,
  disk_used BIGINT NOT NULL DEFAULT 0,
  disk_usage REAL NOT NULL DEFAULT 0,
  uptime VARCHAR(100) NOT NULL DEFAULT '',
  hostname VARCHAR(100) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_host_group_active ON host(group_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS db_instance (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  group_id BIGINT NOT NULL DEFAULT 0,
  name VARCHAR(255) NOT NULL,
  db_type VARCHAR(50) NOT NULL,
  host VARCHAR(255) NOT NULL,
  port INTEGER NOT NULL,
  version VARCHAR(50) NOT NULL DEFAULT '',
  credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL,
  tags VARCHAR(255) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_db_instance_group_active ON db_instance(group_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS terminal_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id BIGINT NOT NULL,
  host_id BIGINT REFERENCES host(id) ON DELETE SET NULL,
  db_instance_id BIGINT REFERENCES db_instance(id) ON DELETE SET NULL,
  session_type VARCHAR(20) NOT NULL,
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ended_at DATETIME,
  duration_seconds BIGINT NOT NULL DEFAULT 0,
  command_count BIGINT NOT NULL DEFAULT 0,
  recording TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS monitor_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain VARCHAR(255) NOT NULL,
  port INTEGER NOT NULL DEFAULT 443,
  request_client VARCHAR(20) NOT NULL DEFAULT 'chrome',
  status VARCHAR(20) NOT NULL DEFAULT 'unknown',
  response_time INTEGER NOT NULL DEFAULT 0,
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE,
  ssl_expiry DATETIME,
  ssl_root_ca VARCHAR(255) NOT NULL DEFAULT '',
  check_interval INTEGER NOT NULL DEFAULT 300,
  check_jitter INTEGER NOT NULL DEFAULT 30,
  enable_ssl BOOLEAN NOT NULL DEFAULT TRUE,
  verify_certificate BOOLEAN NOT NULL DEFAULT TRUE,
  enable_alert BOOLEAN NOT NULL DEFAULT FALSE,
  response_threshold INTEGER NOT NULL DEFAULT 1000,
  ssl_expiry_days INTEGER NOT NULL DEFAULT 30,
  last_check DATETIME,
  next_check DATETIME,
  last_error TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME,
  UNIQUE(domain, port)
);
CREATE INDEX IF NOT EXISTS idx_monitor_domain_next_check ON monitor_domain(next_check);

CREATE TABLE IF NOT EXISTS monitor_check_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER NOT NULL REFERENCES monitor_domain(id) ON DELETE CASCADE,
  status VARCHAR(20) NOT NULL,
  response_time INTEGER NOT NULL DEFAULT 0,
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE,
  ssl_expiry DATETIME,
  status_code INTEGER NOT NULL DEFAULT 0,
  error_message TEXT NOT NULL DEFAULT '',
  checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_monitor_check_history_domain_at ON monitor_check_history(domain_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS monitor_alert_channel (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  channel_type VARCHAR(20) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  config TEXT NOT NULL DEFAULT '{}',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS monitor_alert_receiver (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(100) NOT NULL,
  email VARCHAR(255) NOT NULL DEFAULT '',
  phone VARCHAR(30) NOT NULL DEFAULT '',
  webhook VARCHAR(500) NOT NULL DEFAULT '',
  wechat_id VARCHAR(100) NOT NULL DEFAULT '',
  dingtalk_id VARCHAR(100) NOT NULL DEFAULT '',
  feishu_id VARCHAR(100) NOT NULL DEFAULT '',
  enable_email BOOLEAN NOT NULL DEFAULT TRUE,
  enable_webhook BOOLEAN NOT NULL DEFAULT FALSE,
  enable_wechat BOOLEAN NOT NULL DEFAULT FALSE,
  enable_dingtalk BOOLEAN NOT NULL DEFAULT FALSE,
  enable_feishu BOOLEAN NOT NULL DEFAULT FALSE,
  enable_system_msg BOOLEAN NOT NULL DEFAULT TRUE,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS monitor_alert_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  domain_id INTEGER REFERENCES monitor_domain(id) ON DELETE SET NULL,
  alert_type VARCHAR(50) NOT NULL,
  status VARCHAR(20) NOT NULL,
  message TEXT NOT NULL DEFAULT '',
  channel_type VARCHAR(20) NOT NULL DEFAULT '',
  receiver_names TEXT NOT NULL DEFAULT '',
  error_message TEXT NOT NULL DEFAULT '',
  sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_monitor_alert_log_sent_at ON monitor_alert_log(sent_at DESC);

CREATE TABLE IF NOT EXISTS nginx_source (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT NOT NULL,
  source_type TEXT NOT NULL DEFAULT 'host',
  description TEXT NOT NULL DEFAULT '',
  status INTEGER NOT NULL DEFAULT 1,
  host_id INTEGER NOT NULL DEFAULT 0,
  log_path TEXT NOT NULL DEFAULT '',
  log_format TEXT NOT NULL DEFAULT 'combined',
  collect_interval INTEGER NOT NULL DEFAULT 60,
  retention_days INTEGER NOT NULL DEFAULT 30,
  last_collect_at DATETIME,
  last_collect_logs INTEGER NOT NULL DEFAULT 0,
  last_error TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS nginx_access_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  source_id INTEGER NOT NULL,
  occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  remote_addr TEXT NOT NULL DEFAULT '',
  method TEXT NOT NULL DEFAULT '',
  uri TEXT NOT NULL DEFAULT '',
  status INTEGER NOT NULL DEFAULT 0,
  body_bytes_sent INTEGER NOT NULL DEFAULT 0,
  request_time REAL NOT NULL DEFAULT 0,
  user_agent TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_nginx_access_log_source_time ON nginx_access_log(source_id, occurred_at DESC);

CREATE TABLE IF NOT EXISTS task_template (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, code TEXT NOT NULL UNIQUE,
  description TEXT NOT NULL DEFAULT '', content TEXT NOT NULL DEFAULT '', variables TEXT NOT NULL DEFAULT '',
  category TEXT NOT NULL DEFAULT 'script', platform TEXT NOT NULL DEFAULT '', timeout INTEGER NOT NULL DEFAULT 300,
  sort_order INTEGER NOT NULL DEFAULT 0, status INTEGER NOT NULL DEFAULT 1, created_by INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
);
CREATE TABLE IF NOT EXISTS task_job (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, template_id INTEGER NOT NULL DEFAULT 0,
  task_type TEXT NOT NULL DEFAULT 'manual', status TEXT NOT NULL DEFAULT 'pending', target_hosts TEXT NOT NULL DEFAULT '',
  parameters TEXT NOT NULL DEFAULT '', execute_time DATETIME, result TEXT NOT NULL DEFAULT '', error_message TEXT NOT NULL DEFAULT '',
  created_by INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
);
CREATE TABLE IF NOT EXISTS task_execution (
  id INTEGER PRIMARY KEY AUTOINCREMENT, job_id INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'pending',
  output TEXT NOT NULL DEFAULT '', error_message TEXT NOT NULL DEFAULT '', started_at DATETIME, ended_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_task_job_status ON task_job(status);
CREATE INDEX IF NOT EXISTS idx_task_execution_job ON task_execution(job_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ssl_certificate (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, domain TEXT NOT NULL, san_domains TEXT NOT NULL DEFAULT '',
  source_type TEXT NOT NULL DEFAULT 'manual', ca_provider TEXT NOT NULL DEFAULT 'letsencrypt', certificate TEXT NOT NULL DEFAULT '',
  private_key TEXT NOT NULL DEFAULT '', issuer TEXT NOT NULL DEFAULT '', not_after DATETIME, status TEXT NOT NULL DEFAULT 'pending',
  auto_renew BOOLEAN NOT NULL DEFAULT TRUE, renew_days_before INTEGER NOT NULL DEFAULT 30, last_error TEXT NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
);
CREATE TABLE IF NOT EXISTS ssl_dns_provider (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, provider TEXT NOT NULL, config TEXT NOT NULL DEFAULT '',
  email TEXT NOT NULL DEFAULT '', enabled BOOLEAN NOT NULL DEFAULT TRUE, last_test_at DATETIME,
  last_test_ok BOOLEAN NOT NULL DEFAULT FALSE, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_ssl_certificate_domain ON ssl_certificate(domain);

CREATE TABLE IF NOT EXISTS audit_operation_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '',
  module TEXT NOT NULL DEFAULT '', action TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '', method TEXT NOT NULL DEFAULT 'POST',
  path TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL DEFAULT 200, error_message TEXT NOT NULL DEFAULT '',
  cost_time INTEGER NOT NULL DEFAULT 0, ip TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS audit_login_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '',
  login_type TEXT NOT NULL DEFAULT 'web', status TEXT NOT NULL DEFAULT 'success', ip TEXT NOT NULL DEFAULT '',
  message TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS audit_data_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '',
  table_name TEXT NOT NULL DEFAULT '', record_id INTEGER NOT NULL DEFAULT 0, action TEXT NOT NULL DEFAULT '',
  old_data TEXT NOT NULL DEFAULT '', new_data TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS kubernetes_cluster (
  id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, alias TEXT NOT NULL DEFAULT '', api_endpoint TEXT NOT NULL DEFAULT '',
  kube_config TEXT NOT NULL DEFAULT '', version TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL DEFAULT 0,
  node_count INTEGER NOT NULL DEFAULT 0, pod_count INTEGER NOT NULL DEFAULT 0, region TEXT NOT NULL DEFAULT '',
  provider TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME
);
CREATE TABLE IF NOT EXISTS kubernetes_terminal_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, namespace TEXT NOT NULL DEFAULT '',
  pod_name TEXT NOT NULL DEFAULT '', container_name TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL,
  recording TEXT NOT NULL DEFAULT '', duration INTEGER NOT NULL DEFAULT 0, command_count INTEGER NOT NULL DEFAULT 0,
  status TEXT NOT NULL DEFAULT 'completed', started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_cluster ON kubernetes_terminal_session(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_user ON kubernetes_terminal_session(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_cluster_inspection (
  id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, cluster_name TEXT NOT NULL DEFAULT '',
  status TEXT NOT NULL DEFAULT 'running', score INTEGER NOT NULL DEFAULT 0, check_count INTEGER NOT NULL DEFAULT 0,
  pass_count INTEGER NOT NULL DEFAULT 0, warning_count INTEGER NOT NULL DEFAULT 0, fail_count INTEGER NOT NULL DEFAULT 0,
  duration INTEGER NOT NULL DEFAULT 0, report_data TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL,
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_cluster_inspection_cluster ON kubernetes_cluster_inspection(cluster_id);

CREATE TABLE IF NOT EXISTS kubernetes_user_kube_config (
  id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
  service_account TEXT NOT NULL, namespace TEXT NOT NULL DEFAULT 'gops-auth', is_active INTEGER NOT NULL DEFAULT 1,
  created_by INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, revoked_at DATETIME,
  UNIQUE(cluster_id, user_id, service_account)
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_cluster ON kubernetes_user_kube_config(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_user ON kubernetes_user_kube_config(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_user_role_binding (
  id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL,
  role_name TEXT NOT NULL, role_namespace TEXT NOT NULL DEFAULT '', role_type TEXT NOT NULL, bound_by INTEGER NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(cluster_id, user_id, role_name, role_namespace)
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_cluster ON kubernetes_user_role_binding(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_user ON kubernetes_user_role_binding(user_id);

CREATE TABLE IF NOT EXISTS user_mfa_totp (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL UNIQUE REFERENCES user(id) ON DELETE CASCADE,
  name TEXT NOT NULL DEFAULT '身份验证器',
  secret TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_mfa_totp_user_id_unique ON user_mfa_totp(user_id);

CREATE TABLE IF NOT EXISTS user_mfa_passkey (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE,
  name TEXT NOT NULL DEFAULT '',
  credential_id TEXT NOT NULL UNIQUE,
  credential_json TEXT NOT NULL,
  enabled INTEGER NOT NULL DEFAULT 1,
  direct_login_enabled INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_user_mfa_passkey_user_id ON user_mfa_passkey(user_id);

CREATE TABLE IF NOT EXISTS user_mfa_email (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL UNIQUE REFERENCES user(id) ON DELETE CASCADE,
  mfa_enabled INTEGER NOT NULL DEFAULT 0,
  direct_login_enabled INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE IF EXISTS user_mfa_email;
DROP TABLE IF EXISTS user_mfa_passkey;
DROP TABLE IF EXISTS user_mfa_totp;
DROP TABLE IF EXISTS kubernetes_user_role_binding;
DROP TABLE IF EXISTS kubernetes_user_kube_config;
DROP TABLE IF EXISTS kubernetes_cluster_inspection;
DROP TABLE IF EXISTS kubernetes_terminal_session;
DROP TABLE IF EXISTS kubernetes_cluster;
DROP TABLE IF EXISTS audit_data_log;
DROP TABLE IF EXISTS audit_login_log;
DROP TABLE IF EXISTS audit_operation_log;
DROP TABLE IF EXISTS ssl_dns_provider;
DROP TABLE IF EXISTS ssl_certificate;
DROP TABLE IF EXISTS task_execution;
DROP TABLE IF EXISTS task_job;
DROP TABLE IF EXISTS task_template;
DROP TABLE IF EXISTS nginx_access_log;
DROP TABLE IF EXISTS nginx_source;
DROP TABLE IF EXISTS monitor_alert_log;
DROP TABLE IF EXISTS monitor_alert_receiver;
DROP TABLE IF EXISTS monitor_alert_channel;
DROP TABLE IF EXISTS monitor_check_history;
DROP TABLE IF EXISTS monitor_domain;
DROP TABLE IF EXISTS terminal_session;
DROP TABLE IF EXISTS db_instance;
DROP TABLE IF EXISTS host;
DROP TABLE IF EXISTS cloud_account;
DROP TABLE IF EXISTS credential;
DROP TABLE IF EXISTS asset_group;
DROP TABLE IF EXISTS system_config;
DROP TABLE IF EXISTS organization_user;
DROP TABLE IF EXISTS organization;
DROP TABLE IF EXISTS user_role;
DROP TABLE IF EXISTS user;
DROP TABLE IF EXISTS role;
