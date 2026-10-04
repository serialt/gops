-- +goose Up
-- Gops SQLite 最终基线结构。历史建表、字段调整与数据修复迁移已合并至此文件。

CREATE TABLE IF NOT EXISTS role (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 角色ID */
  code VARCHAR(64) NOT NULL, /* 角色编码 */
  name VARCHAR(64) NOT NULL, /* 角色名称 */
  description VARCHAR(255), /* 角色说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_role_code_active ON role(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS user (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 用户ID */
  username VARCHAR(64) NOT NULL, /* 用户名 */
  password_hash VARCHAR(255) NOT NULL, /* 密码哈希 */
  nickname VARCHAR(64) NOT NULL, /* 用户昵称 */
  email VARCHAR(255) NOT NULL DEFAULT '', /* 邮箱地址 */
  status BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_username_active ON user(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_email_active ON user(email) WHERE deleted_at IS NULL AND email <> '';

CREATE TABLE IF NOT EXISTS user_role (
  user_id INTEGER NOT NULL, /* 用户ID */
  role_id INTEGER NOT NULL, /* 角色ID */
  PRIMARY KEY (user_id, role_id)
);
CREATE INDEX IF NOT EXISTS idx_user_role_role_id ON user_role(role_id);

CREATE TABLE IF NOT EXISTS organization (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 组织ID */
  parent_id INTEGER REFERENCES organization(id) ON DELETE RESTRICT, /* 上级组织ID */
  name VARCHAR(64) NOT NULL, /* 组织名称 */
  code VARCHAR(64) NOT NULL, /* 组织编码 */
  manager VARCHAR(64) NOT NULL DEFAULT '', /* 负责人 */
  status BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME, /* 删除时间 */
  sort_order INTEGER NOT NULL DEFAULT 0 /* 排序值 */
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_organization_code_active ON organization(code) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_organization_parent_id ON organization(parent_id);

CREATE TABLE IF NOT EXISTS organization_user (
  organization_id INTEGER NOT NULL REFERENCES organization(id) ON DELETE CASCADE, /* 组织ID */
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE, /* 用户ID */
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
  config_key VARCHAR(128) PRIMARY KEY, /* 配置键 */
  config_value TEXT NOT NULL DEFAULT '', /* 配置值 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
INSERT INTO system_config(config_key, config_value) VALUES ('system_name', 'Gops') ON CONFLICT(config_key) DO NOTHING;
INSERT INTO system_config(config_key, config_value) VALUES ('system_description', '') ON CONFLICT(config_key) DO NOTHING;

CREATE TABLE IF NOT EXISTS asset_group (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 资产分组ID */
  name VARCHAR(100) NOT NULL, /* 分组名称 */
  code VARCHAR(50) NOT NULL, /* 分组编码 */
  parent_id BIGINT NOT NULL DEFAULT 0, /* 上级分组ID */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 分组说明 */
  sort_order INTEGER NOT NULL DEFAULT 0, /* 排序值 */
  status BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_group_code_active ON asset_group(code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS credential (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 凭据ID */
  name VARCHAR(100) NOT NULL, /* 凭据名称 */
  credential_type VARCHAR(20) NOT NULL, /* 凭据类型 */
  username VARCHAR(100) NOT NULL DEFAULT '', /* 登录用户名 */
  password TEXT NOT NULL DEFAULT '', /* 加密密码 */
  private_key TEXT NOT NULL DEFAULT '', /* 加密私钥 */
  passphrase TEXT NOT NULL DEFAULT '', /* 私钥口令 */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 凭据说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);

CREATE TABLE IF NOT EXISTS cloud_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 云账号ID */
  name VARCHAR(100) NOT NULL, /* 云账号名称 */
  provider VARCHAR(50) NOT NULL, /* 云服务商 */
  access_key TEXT NOT NULL DEFAULT '', /* 访问密钥 */
  secret_key TEXT NOT NULL DEFAULT '', /* 加密访问密钥 */
  region VARCHAR(100) NOT NULL DEFAULT '', /* 默认区域 */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 账号说明 */
  status BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);

CREATE TABLE IF NOT EXISTS host (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 主机ID */
  name VARCHAR(100) NOT NULL, /* 主机名称 */
  group_id BIGINT REFERENCES asset_group(id) ON DELETE SET NULL, /* 资产分组ID */
  host_type VARCHAR(20) NOT NULL DEFAULT 'self', /* 主机类型 */
  ip VARCHAR(255) NOT NULL, /* 主机地址 */
  port INTEGER NOT NULL DEFAULT 22, /* SSH端口 */
  credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL, /* 凭据ID */
  cloud_account_id BIGINT REFERENCES cloud_account(id) ON DELETE SET NULL, /* 云账号ID */
  tags VARCHAR(500) NOT NULL DEFAULT '', /* 主机标签 */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 主机说明 */
  status SMALLINT NOT NULL DEFAULT -1, /* 在线状态 */
  last_seen_at DATETIME, /* 最后在线时间 */
  os VARCHAR(100) NOT NULL DEFAULT '', /* 操作系统 */
  kernel VARCHAR(100) NOT NULL DEFAULT '', /* 内核版本 */
  arch VARCHAR(50) NOT NULL DEFAULT '', /* 系统架构 */
  cpu_info TEXT NOT NULL DEFAULT '', /* CPU信息 */
  cpu_cores INTEGER NOT NULL DEFAULT 0, /* CPU核心数 */
  cpu_usage REAL NOT NULL DEFAULT 0, /* CPU使用率 */
  memory_total BIGINT NOT NULL DEFAULT 0, /* 内存总量 */
  memory_used BIGINT NOT NULL DEFAULT 0, /* 已用内存 */
  memory_usage REAL NOT NULL DEFAULT 0, /* 内存使用率 */
  disk_total BIGINT NOT NULL DEFAULT 0, /* 磁盘总量 */
  disk_used BIGINT NOT NULL DEFAULT 0, /* 已用磁盘 */
  disk_usage REAL NOT NULL DEFAULT 0, /* 磁盘使用率 */
  uptime VARCHAR(100) NOT NULL DEFAULT '', /* 运行时长 */
  hostname VARCHAR(100) NOT NULL DEFAULT '', /* 系统主机名 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_host_group_active ON host(group_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS db_instance (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 数据库实例ID */
  group_id BIGINT NOT NULL DEFAULT 0, /* 资产分组ID */
  name VARCHAR(255) NOT NULL, /* 实例名称 */
  db_type VARCHAR(50) NOT NULL, /* 数据库类型 */
  host VARCHAR(255) NOT NULL, /* 数据库地址 */
  port INTEGER NOT NULL, /* 数据库端口 */
  version VARCHAR(50) NOT NULL DEFAULT '', /* 数据库版本 */
  credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL, /* 凭据ID */
  tags VARCHAR(255) NOT NULL DEFAULT '', /* 实例标签 */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 实例说明 */
  status BOOLEAN NOT NULL DEFAULT TRUE, /* 连接状态 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_db_instance_group_active ON db_instance(group_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS terminal_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 终端会话ID */
  user_id BIGINT NOT NULL, /* 用户ID */
  host_id BIGINT REFERENCES host(id) ON DELETE SET NULL, /* 主机ID */
  db_instance_id BIGINT REFERENCES db_instance(id) ON DELETE SET NULL, /* 数据库实例ID */
  session_type VARCHAR(20) NOT NULL, /* 会话类型 */
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 开始时间 */
  ended_at DATETIME, /* 结束时间 */
  duration_seconds BIGINT NOT NULL DEFAULT 0, /* 会话时长秒数 */
  command_count BIGINT NOT NULL DEFAULT 0, /* 命令数量 */
  recording TEXT NOT NULL DEFAULT '', /* 会话录像 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);

CREATE TABLE IF NOT EXISTS monitor_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 域名监控ID */
  domain VARCHAR(255) NOT NULL, /* 域名 */
  port INTEGER NOT NULL DEFAULT 443, /* 访问端口 */
  request_client VARCHAR(20) NOT NULL DEFAULT 'chrome', /* 请求客户端类型 */
  status VARCHAR(20) NOT NULL DEFAULT 'unknown', /* 监控状态 */
  response_time INTEGER NOT NULL DEFAULT 0, /* 响应耗时毫秒 */
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE, /* SSL证书是否有效 */
  ssl_expiry DATETIME, /* SSL证书过期时间 */
  ssl_root_ca VARCHAR(255) NOT NULL DEFAULT '', /* SSL根证书机构 */
  check_interval INTEGER NOT NULL DEFAULT 300, /* 检查间隔秒数 */
  check_jitter INTEGER NOT NULL DEFAULT 30, /* 检查抖动秒数 */
  enable_ssl BOOLEAN NOT NULL DEFAULT TRUE, /* 是否检查SSL */
  verify_certificate BOOLEAN NOT NULL DEFAULT TRUE, /* 是否验证证书 */
  enable_alert BOOLEAN NOT NULL DEFAULT FALSE, /* 是否启用告警 */
  response_threshold INTEGER NOT NULL DEFAULT 1000, /* 响应时间告警阈值 */
  ssl_expiry_days INTEGER NOT NULL DEFAULT 30, /* 证书到期预警天数 */
  last_check DATETIME, /* 上次检查时间 */
  next_check DATETIME, /* 下次检查时间 */
  last_error TEXT NOT NULL DEFAULT '', /* 最近错误 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_monitor_domain_next_check ON monitor_domain(next_check);
CREATE UNIQUE INDEX IF NOT EXISTS idx_monitor_domain_address_active ON monitor_domain(domain, port) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS monitor_check_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 检查记录ID */
  domain_id INTEGER NOT NULL REFERENCES monitor_domain(id) ON DELETE CASCADE, /* 域名监控ID */
  status VARCHAR(20) NOT NULL, /* 检查状态 */
  response_time INTEGER NOT NULL DEFAULT 0, /* 响应耗时毫秒 */
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE, /* SSL证书是否有效 */
  ssl_expiry DATETIME, /* SSL证书过期时间 */
  status_code INTEGER NOT NULL DEFAULT 0, /* HTTP状态码 */
  error_message TEXT NOT NULL DEFAULT '', /* 错误信息 */
  checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 检查时间 */
);
CREATE INDEX IF NOT EXISTS idx_monitor_check_history_domain_at ON monitor_check_history(domain_id, checked_at DESC);

CREATE TABLE IF NOT EXISTS monitor_alert_channel (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 告警渠道ID */
  name VARCHAR(100) NOT NULL, /* 渠道名称 */
  channel_type VARCHAR(20) NOT NULL, /* 渠道类型 */
  enabled BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  config TEXT NOT NULL DEFAULT '{}', /* 渠道配置JSON */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);

CREATE TABLE IF NOT EXISTS monitor_alert_receiver (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 告警接收人ID */
  name VARCHAR(100) NOT NULL, /* 接收人名称 */
  email VARCHAR(255) NOT NULL DEFAULT '', /* 邮箱地址 */
  phone VARCHAR(30) NOT NULL DEFAULT '', /* 手机号码 */
  webhook VARCHAR(500) NOT NULL DEFAULT '', /* Webhook地址 */
  wechat_id VARCHAR(100) NOT NULL DEFAULT '', /* 企业微信接收标识 */
  dingtalk_id VARCHAR(100) NOT NULL DEFAULT '', /* 钉钉接收标识 */
  feishu_id VARCHAR(100) NOT NULL DEFAULT '', /* 飞书接收标识 */
  enable_email BOOLEAN NOT NULL DEFAULT TRUE, /* 是否接收邮件告警 */
  enable_webhook BOOLEAN NOT NULL DEFAULT FALSE, /* 是否接收Webhook告警 */
  enable_wechat BOOLEAN NOT NULL DEFAULT FALSE, /* 是否接收企业微信告警 */
  enable_dingtalk BOOLEAN NOT NULL DEFAULT FALSE, /* 是否接收钉钉告警 */
  enable_feishu BOOLEAN NOT NULL DEFAULT FALSE, /* 是否接收飞书告警 */
  enable_system_msg BOOLEAN NOT NULL DEFAULT TRUE, /* 是否接收系统消息 */
  enabled BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);

CREATE TABLE IF NOT EXISTS monitor_alert_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 告警日志ID */
  domain_id INTEGER REFERENCES monitor_domain(id) ON DELETE SET NULL, /* 域名监控ID */
  alert_type VARCHAR(50) NOT NULL, /* 告警类型 */
  status VARCHAR(20) NOT NULL, /* 发送状态 */
  message TEXT NOT NULL DEFAULT '', /* 告警内容 */
  channel_type VARCHAR(20) NOT NULL DEFAULT '', /* 发送渠道类型 */
  receiver_names TEXT NOT NULL DEFAULT '', /* 接收人名称列表 */
  error_message TEXT NOT NULL DEFAULT '', /* 发送错误信息 */
  sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 发送时间 */
);
CREATE INDEX IF NOT EXISTS idx_monitor_alert_log_sent_at ON monitor_alert_log(sent_at DESC);

CREATE TABLE IF NOT EXISTS task_template (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 任务模板ID */
  name TEXT NOT NULL, /* 模板名称 */
  code TEXT NOT NULL, /* 模板编码 */
  description TEXT NOT NULL DEFAULT '', /* 模板说明 */
  content TEXT NOT NULL DEFAULT '', /* 模板内容 */
  variables TEXT NOT NULL DEFAULT '', /* 模板变量JSON */
  category TEXT NOT NULL DEFAULT 'script', /* 模板分类 */
  platform TEXT NOT NULL DEFAULT '', /* 适用平台 */
  timeout INTEGER NOT NULL DEFAULT 300, /* 超时时间秒数 */
  sort_order INTEGER NOT NULL DEFAULT 0, /* 排序值 */
  status INTEGER NOT NULL DEFAULT 1, /* 是否启用 */
  created_by INTEGER NOT NULL DEFAULT 0, /* 创建用户ID */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_task_template_code_active ON task_template(code) WHERE deleted_at IS NULL;
CREATE TABLE IF NOT EXISTS task_job (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 任务ID */
  name TEXT NOT NULL, /* 任务名称 */
  template_id INTEGER NOT NULL DEFAULT 0, /* 任务模板ID */
  task_type TEXT NOT NULL DEFAULT 'manual', /* 任务类型 */
  status TEXT NOT NULL DEFAULT 'pending', /* 执行状态 */
  target_hosts TEXT NOT NULL DEFAULT '', /* 目标主机列表 */
  parameters TEXT NOT NULL DEFAULT '', /* 执行参数JSON */
  execute_time DATETIME, /* 计划执行时间 */
  result TEXT NOT NULL DEFAULT '', /* 执行结果 */
  error_message TEXT NOT NULL DEFAULT '', /* 错误信息 */
  created_by INTEGER NOT NULL DEFAULT 0, /* 创建用户ID */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE IF NOT EXISTS task_execution (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 执行记录ID */
  job_id INTEGER NOT NULL, /* 任务ID */
  status TEXT NOT NULL DEFAULT 'pending', /* 执行状态 */
  output TEXT NOT NULL DEFAULT '', /* 执行输出 */
  error_message TEXT NOT NULL DEFAULT '', /* 错误信息 */
  started_at DATETIME, /* 开始时间 */
  ended_at DATETIME, /* 结束时间 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 创建时间 */
);
CREATE INDEX IF NOT EXISTS idx_task_job_status ON task_job(status);
CREATE INDEX IF NOT EXISTS idx_task_execution_job ON task_execution(job_id, created_at DESC);

CREATE TABLE IF NOT EXISTS ssl_certificate (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* SSL证书ID */
  name TEXT NOT NULL, /* 证书名称 */
  domain TEXT NOT NULL, /* 主域名 */
  san_domains TEXT NOT NULL DEFAULT '', /* 备用域名列表 */
  source_type TEXT NOT NULL DEFAULT 'manual', /* 证书来源 */
  ca_provider TEXT NOT NULL DEFAULT 'letsencrypt', /* 证书颁发机构 */
  certificate TEXT NOT NULL DEFAULT '', /* 证书内容 */
  private_key TEXT NOT NULL DEFAULT '', /* 私钥内容 */
  issuer TEXT NOT NULL DEFAULT '', /* 签发者 */
  not_after DATETIME, /* 过期时间 */
  status TEXT NOT NULL DEFAULT 'pending', /* 证书状态 */
  auto_renew BOOLEAN NOT NULL DEFAULT TRUE, /* 是否自动续期 */
  renew_days_before INTEGER NOT NULL DEFAULT 30, /* 提前续期天数 */
  last_error TEXT NOT NULL DEFAULT '', /* 最近错误 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE IF NOT EXISTS ssl_dns_provider (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* DNS服务商配置ID */
  name TEXT NOT NULL, /* 配置名称 */
  provider TEXT NOT NULL, /* DNS服务商 */
  config TEXT NOT NULL DEFAULT '', /* 服务商配置JSON */
  email TEXT NOT NULL DEFAULT '', /* 联系邮箱 */
  enabled BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  last_test_at DATETIME, /* 最近测试时间 */
  last_test_ok BOOLEAN NOT NULL DEFAULT FALSE, /* 最近测试是否成功 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX IF NOT EXISTS idx_ssl_certificate_domain ON ssl_certificate(domain);

CREATE TABLE IF NOT EXISTS audit_operation_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 操作日志ID */
  user_id INTEGER NOT NULL DEFAULT 0, /* 操作用户ID */
  username TEXT NOT NULL DEFAULT '', /* 操作用户名 */
  module TEXT NOT NULL DEFAULT '', /* 业务模块 */
  action TEXT NOT NULL DEFAULT '', /* 操作类型 */
  description TEXT NOT NULL DEFAULT '', /* 操作描述 */
  method TEXT NOT NULL DEFAULT 'POST', /* 请求方法 */
  path TEXT NOT NULL DEFAULT '', /* 请求路径 */
  status INTEGER NOT NULL DEFAULT 200, /* 响应状态码 */
  error_message TEXT NOT NULL DEFAULT '', /* 错误信息 */
  cost_time INTEGER NOT NULL DEFAULT 0, /* 请求耗时毫秒 */
  ip TEXT NOT NULL DEFAULT '', /* 客户端IP */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 创建时间 */
);
CREATE TABLE IF NOT EXISTS audit_login_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 登录日志ID */
  user_id INTEGER NOT NULL DEFAULT 0, /* 登录用户ID */
  username TEXT NOT NULL DEFAULT '', /* 登录账号 */
  login_type TEXT NOT NULL DEFAULT 'web', /* 登录类型 */
  status TEXT NOT NULL DEFAULT 'success', /* 登录状态 */
  ip TEXT NOT NULL DEFAULT '', /* 客户端IP */
  message TEXT NOT NULL DEFAULT '', /* 登录结果说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 创建时间 */
);
CREATE TABLE IF NOT EXISTS audit_data_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 数据审计日志ID */
  user_id INTEGER NOT NULL DEFAULT 0, /* 操作用户ID */
  username TEXT NOT NULL DEFAULT '', /* 操作用户名 */
  table_name TEXT NOT NULL DEFAULT '', /* 数据表名称 */
  record_id INTEGER NOT NULL DEFAULT 0, /* 数据记录ID */
  action TEXT NOT NULL DEFAULT '', /* 数据操作类型 */
  old_data TEXT NOT NULL DEFAULT '', /* 变更前数据JSON */
  new_data TEXT NOT NULL DEFAULT '', /* 变更后数据JSON */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 创建时间 */
);

CREATE TABLE IF NOT EXISTS kubernetes_cluster (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 集群ID */
  name TEXT NOT NULL, /* 集群名称 */
  alias TEXT NOT NULL DEFAULT '', /* 集群别名 */
  api_endpoint TEXT NOT NULL DEFAULT '', /* API服务地址 */
  kube_config TEXT NOT NULL DEFAULT '', /* KubeConfig内容 */
  version TEXT NOT NULL DEFAULT '', /* Kubernetes版本 */
  status INTEGER NOT NULL DEFAULT 0, /* 集群状态 */
  node_count INTEGER NOT NULL DEFAULT 0, /* 节点数量 */
  pod_count INTEGER NOT NULL DEFAULT 0, /* Pod数量 */
  region TEXT NOT NULL DEFAULT '', /* 所属区域 */
  provider TEXT NOT NULL DEFAULT '', /* 集群提供商 */
  description TEXT NOT NULL DEFAULT '', /* 集群说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE IF NOT EXISTS kubernetes_terminal_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* Kubernetes终端会话ID */
  cluster_id INTEGER NOT NULL, /* 集群ID */
  namespace TEXT NOT NULL DEFAULT '', /* 命名空间 */
  pod_name TEXT NOT NULL DEFAULT '', /* Pod名称 */
  container_name TEXT NOT NULL DEFAULT '', /* 容器名称 */
  user_id INTEGER NOT NULL, /* 用户ID */
  recording TEXT NOT NULL DEFAULT '', /* 会话录像 */
  duration INTEGER NOT NULL DEFAULT 0, /* 会话时长秒数 */
  command_count INTEGER NOT NULL DEFAULT 0, /* 命令数量 */
  status TEXT NOT NULL DEFAULT 'completed', /* 会话状态 */
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 开始时间 */
  ended_at DATETIME, /* 结束时间 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_cluster ON kubernetes_terminal_session(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_terminal_session_user ON kubernetes_terminal_session(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_cluster_inspection (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 集群巡检ID */
  cluster_id INTEGER NOT NULL, /* 集群ID */
  cluster_name TEXT NOT NULL DEFAULT '', /* 集群名称快照 */
  status TEXT NOT NULL DEFAULT 'running', /* 巡检状态 */
  score INTEGER NOT NULL DEFAULT 0, /* 巡检得分 */
  check_count INTEGER NOT NULL DEFAULT 0, /* 检查项总数 */
  pass_count INTEGER NOT NULL DEFAULT 0, /* 通过数量 */
  warning_count INTEGER NOT NULL DEFAULT 0, /* 警告数量 */
  fail_count INTEGER NOT NULL DEFAULT 0, /* 失败数量 */
  duration INTEGER NOT NULL DEFAULT 0, /* 巡检耗时秒数 */
  report_data TEXT NOT NULL DEFAULT '', /* 巡检报告JSON */
  user_id INTEGER NOT NULL, /* 执行用户ID */
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 开始时间 */
  ended_at DATETIME, /* 结束时间 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_cluster_inspection_cluster ON kubernetes_cluster_inspection(cluster_id);

CREATE TABLE IF NOT EXISTS kubernetes_user_kube_config (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 用户KubeConfig ID */
  cluster_id INTEGER NOT NULL, /* 集群ID */
  user_id INTEGER NOT NULL, /* 用户ID */
  service_account TEXT NOT NULL, /* ServiceAccount名称 */
  namespace TEXT NOT NULL DEFAULT 'gops-auth', /* ServiceAccount命名空间 */
  is_active INTEGER NOT NULL DEFAULT 1, /* 是否有效 */
  created_by INTEGER NOT NULL, /* 创建用户ID */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  revoked_at DATETIME, /* 撤销时间 */
  UNIQUE(cluster_id, user_id, service_account)
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_cluster ON kubernetes_user_kube_config(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_kube_config_user ON kubernetes_user_kube_config(user_id);

CREATE TABLE IF NOT EXISTS kubernetes_user_role_binding (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 用户角色绑定ID */
  cluster_id INTEGER NOT NULL, /* 集群ID */
  user_id INTEGER NOT NULL, /* 用户ID */
  role_name TEXT NOT NULL, /* 角色名称 */
  role_namespace TEXT NOT NULL DEFAULT '', /* 角色命名空间 */
  role_type TEXT NOT NULL, /* 角色类型 */
  bound_by INTEGER NOT NULL, /* 绑定操作用户ID */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  UNIQUE(cluster_id, user_id, role_name, role_namespace)
);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_cluster ON kubernetes_user_role_binding(cluster_id);
CREATE INDEX IF NOT EXISTS idx_kubernetes_user_role_binding_user ON kubernetes_user_role_binding(user_id);

-- +goose Down
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
