-- +goose Up
-- Gops SQLite 最终基线结构，由截至 2026-10-08 的全部历史迁移合并生成。
PRAGMA foreign_keys = ON;

CREATE TABLE "user_role" (
    "user_id" INTEGER NOT NULL, -- 用户主键
    "role_id" INTEGER NOT NULL, -- 角色主键
    PRIMARY KEY ("user_id", "role_id")
 
);
CREATE TABLE "organization" (
  "id" INTEGER PRIMARY KEY AUTOINCREMENT, -- 组织主键
  "parent_id" INTEGER REFERENCES "organization"("id") ON DELETE RESTRICT, -- 上级组织主键，根组织为空
  "name" VARCHAR(64) NOT NULL, -- 组织名称
  "code" VARCHAR(64) NOT NULL, -- 组织编码，系统内唯一
  "manager" VARCHAR(64) NOT NULL DEFAULT '', -- 组织负责人
  "status" BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  "deleted_at" DATETIME, -- 软删除时间
  "sort_order" INTEGER NOT NULL DEFAULT 0 -- 同级组织排序值
);
CREATE TABLE "organization_user" (
  "organization_id" INTEGER NOT NULL REFERENCES "organization"("id") ON DELETE CASCADE, -- 组织主键
  "user_id" INTEGER NOT NULL REFERENCES "user"("id") ON DELETE CASCADE, -- 用户主键
  PRIMARY KEY ("organization_id", "user_id")
);
CREATE TABLE "system_config" (
  "config_key" VARCHAR(128) PRIMARY KEY, -- 配置键
  "config_value" TEXT NOT NULL DEFAULT '', -- 配置值
  "created_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  "updated_at" DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 更新时间
, "deleted_at" DATETIME);
CREATE TABLE asset_group (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 资产分组主键
  name VARCHAR(100) NOT NULL, -- 分组名称
  code VARCHAR(50) NOT NULL, -- 分组编码
  parent_id BIGINT NOT NULL DEFAULT 0, -- 上级分组主键，0 表示顶级分组
  description VARCHAR(500) NOT NULL DEFAULT '', -- 分组说明
  sort_order INTEGER NOT NULL DEFAULT 0, -- 同级排序值
  status BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);
CREATE TABLE credential (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 凭证主键
  name VARCHAR(100) NOT NULL, -- 凭证名称
  credential_type VARCHAR(20) NOT NULL, -- 凭证类型
  username VARCHAR(100) NOT NULL DEFAULT '', -- 登录用户名
  password TEXT NOT NULL DEFAULT '', -- 加密密码
  private_key TEXT NOT NULL DEFAULT '', -- 加密私钥
  passphrase TEXT NOT NULL DEFAULT '', -- 加密私钥口令
  description VARCHAR(500) NOT NULL DEFAULT '', -- 凭证说明
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);
CREATE TABLE cloud_account (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 云账号主键
  name VARCHAR(100) NOT NULL, -- 账号名称
  provider VARCHAR(50) NOT NULL, -- 云服务商
  access_key TEXT NOT NULL DEFAULT '', -- 加密访问密钥
  secret_key TEXT NOT NULL DEFAULT '', -- 加密访问密钥密码
  region VARCHAR(100) NOT NULL DEFAULT '', -- 默认区域
  description VARCHAR(500) NOT NULL DEFAULT '', -- 账号说明
  status BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);
CREATE TABLE host (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 主机主键
  name VARCHAR(100) NOT NULL, -- 主机名称
  group_id BIGINT, -- 所属资产分组主键
  host_type VARCHAR(20) NOT NULL DEFAULT 'self', -- 主机类型
  ip VARCHAR(255) NOT NULL, -- 主机地址
  credential_id BIGINT, -- 关联凭证主键
  cloud_account_id BIGINT, -- 关联云账号主键
  tags VARCHAR(500) NOT NULL DEFAULT '', -- 主机标签
  description VARCHAR(500) NOT NULL DEFAULT '', -- 主机说明
  status SMALLINT NOT NULL DEFAULT -1, -- 连接状态
  last_seen_at DATETIME, -- 最后连接时间
  os VARCHAR(100) NOT NULL DEFAULT '', -- 操作系统
  kernel VARCHAR(100) NOT NULL DEFAULT '', -- 内核版本
  arch VARCHAR(50) NOT NULL DEFAULT '', -- 系统架构
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME, cpu_info TEXT NOT NULL DEFAULT '', cpu_cores INTEGER NOT NULL DEFAULT 0, cpu_usage REAL NOT NULL DEFAULT 0, memory_total BIGINT NOT NULL DEFAULT 0, memory_used BIGINT NOT NULL DEFAULT 0, memory_usage REAL NOT NULL DEFAULT 0, disk_total BIGINT NOT NULL DEFAULT 0, disk_used BIGINT NOT NULL DEFAULT 0, disk_usage REAL NOT NULL DEFAULT 0, uptime VARCHAR(100) NOT NULL DEFAULT '', hostname VARCHAR(100) NOT NULL DEFAULT '', vnc_credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL /* VNC凭据ID */, port INTEGER /* SSH端口 */, vnc_port INTEGER /* VNC端口 */, -- 软删除时间
  CONSTRAINT fk_host_group FOREIGN KEY (group_id) REFERENCES asset_group(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_cloud_account FOREIGN KEY (cloud_account_id) REFERENCES cloud_account(id) ON DELETE SET NULL
);
CREATE TABLE db_instance (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 数据库主键
  name VARCHAR(255) NOT NULL, -- 数据库名称
  db_type VARCHAR(50) NOT NULL, -- 数据库类型
  host VARCHAR(255) NOT NULL, -- 数据库地址
  port INTEGER NOT NULL, -- 数据库端口
  version VARCHAR(50) NOT NULL DEFAULT '', -- 数据库版本
  credential_id BIGINT, -- 关联凭证主键
  tags VARCHAR(255) NOT NULL DEFAULT '', -- 数据库标签
  description VARCHAR(500) NOT NULL DEFAULT '', -- 数据库说明
  status BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME, group_id BIGINT NOT NULL DEFAULT 0, -- 软删除时间
  CONSTRAINT fk_db_instance_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL
);
CREATE TABLE terminal_session (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 会话主键
  user_id BIGINT NOT NULL, -- 操作用户主键
  host_id BIGINT, -- 关联主机主键
  db_instance_id BIGINT, -- 关联数据库主键
  session_type VARCHAR(20) NOT NULL, -- 会话类型
  started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 开始时间
  ended_at DATETIME, -- 结束时间
  duration_seconds BIGINT NOT NULL DEFAULT 0, -- 会话时长（秒）
  command_count BIGINT NOT NULL DEFAULT 0, -- 命令数量
  recording TEXT NOT NULL DEFAULT '', -- 会话录制内容
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME, -- 软删除时间
  CONSTRAINT fk_terminal_session_host FOREIGN KEY (host_id) REFERENCES host(id) ON DELETE SET NULL,
  CONSTRAINT fk_terminal_session_db FOREIGN KEY (db_instance_id) REFERENCES db_instance(id) ON DELETE SET NULL
);
CREATE TABLE monitor_check_history (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 历史记录主键
  domain_id INTEGER NOT NULL REFERENCES monitor_domain(id) ON DELETE CASCADE, -- 域名监控主键
  status VARCHAR(20) NOT NULL, -- 检查状态
  response_time INTEGER NOT NULL DEFAULT 0, -- 响应耗时（毫秒）
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE, -- SSL 是否有效
  ssl_expiry DATETIME, -- SSL 到期时间
  status_code INTEGER NOT NULL DEFAULT 0, -- HTTP 状态码
  error_message TEXT NOT NULL DEFAULT '', -- 错误信息
  checked_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 检查时间
);
CREATE TABLE monitor_alert_channel (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 通道主键
  name VARCHAR(100) NOT NULL, -- 通道名称
  channel_type VARCHAR(20) NOT NULL, -- 通道类型：email、webhook、wechat、dingtalk、feishu
  enabled BOOLEAN NOT NULL DEFAULT TRUE, -- 是否启用
  config TEXT NOT NULL DEFAULT '{}', -- 通道 JSON 配置
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);
CREATE TABLE monitor_alert_receiver (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 接收人主键
  name VARCHAR(100) NOT NULL, -- 接收人名称
  email VARCHAR(255) NOT NULL DEFAULT '', -- 邮箱地址
  phone VARCHAR(30) NOT NULL DEFAULT '', -- 手机号码
  webhook VARCHAR(500) NOT NULL DEFAULT '', -- 个人 Webhook 地址
  enabled BOOLEAN NOT NULL DEFAULT TRUE, -- 是否启用
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
, wechat_id VARCHAR(100) NOT NULL DEFAULT '', dingtalk_id VARCHAR(100) NOT NULL DEFAULT '', feishu_id VARCHAR(100) NOT NULL DEFAULT '', enable_email BOOLEAN NOT NULL DEFAULT TRUE, enable_webhook BOOLEAN NOT NULL DEFAULT FALSE, enable_wechat BOOLEAN NOT NULL DEFAULT FALSE, enable_dingtalk BOOLEAN NOT NULL DEFAULT FALSE, enable_feishu BOOLEAN NOT NULL DEFAULT FALSE, enable_system_msg BOOLEAN NOT NULL DEFAULT TRUE);
CREATE TABLE monitor_alert_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 告警日志主键
  domain_id INTEGER REFERENCES monitor_domain(id) ON DELETE SET NULL, -- 域名监控主键
  alert_type VARCHAR(50) NOT NULL, -- 告警类型
  status VARCHAR(20) NOT NULL, -- 发送状态：success、failed
  message TEXT NOT NULL DEFAULT '', -- 告警内容
  channel_type VARCHAR(20) NOT NULL DEFAULT '', -- 发送通道类型
  error_message TEXT NOT NULL DEFAULT '', -- 发送失败原因
  sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 发送时间
);
CREATE TABLE nginx_source (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 主键
  name TEXT NOT NULL, -- 数据源名称
  source_type TEXT NOT NULL DEFAULT 'host', -- 数据源类型
  description TEXT NOT NULL DEFAULT '', -- 说明
  status INTEGER NOT NULL DEFAULT 1, -- 状态：0禁用，1启用
  host_id INTEGER NOT NULL DEFAULT 0, -- 主机 ID
  log_path TEXT NOT NULL DEFAULT '', -- 日志路径
  log_format TEXT NOT NULL DEFAULT 'combined', -- 日志格式
  collect_interval INTEGER NOT NULL DEFAULT 60, -- 采集间隔秒
  retention_days INTEGER NOT NULL DEFAULT 30, -- 保留天数
  last_collect_at DATETIME, -- 最后采集时间
  last_collect_logs INTEGER NOT NULL DEFAULT 0, -- 最后采集日志数
  last_error TEXT NOT NULL DEFAULT '', -- 最后错误
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 逻辑删除时间
);
CREATE TABLE nginx_access_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 主键
  source_id INTEGER NOT NULL, -- 数据源 ID
  occurred_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 请求时间
  remote_addr TEXT NOT NULL DEFAULT '', -- 客户端 IP
  method TEXT NOT NULL DEFAULT '', -- 请求方法
  uri TEXT NOT NULL DEFAULT '', -- 请求路径
  status INTEGER NOT NULL DEFAULT 0, -- HTTP 状态码
  body_bytes_sent INTEGER NOT NULL DEFAULT 0, -- 响应字节
  request_time REAL NOT NULL DEFAULT 0, -- 请求耗时秒
  user_agent TEXT NOT NULL DEFAULT '', -- 用户代理
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 创建时间
);
CREATE TABLE task_job (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, template_id INTEGER NOT NULL DEFAULT 0, task_type TEXT NOT NULL DEFAULT 'manual', status TEXT NOT NULL DEFAULT 'pending', target_hosts TEXT NOT NULL DEFAULT '', parameters TEXT NOT NULL DEFAULT '', execute_time DATETIME, result TEXT NOT NULL DEFAULT '', error_message TEXT NOT NULL DEFAULT '', created_by INTEGER NOT NULL DEFAULT 0, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME);
CREATE TABLE task_execution (id INTEGER PRIMARY KEY AUTOINCREMENT, job_id INTEGER NOT NULL, status TEXT NOT NULL DEFAULT 'pending', output TEXT NOT NULL DEFAULT '', error_message TEXT NOT NULL DEFAULT '', started_at DATETIME, ended_at DATETIME, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE ssl_certificate (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, domain TEXT NOT NULL, san_domains TEXT NOT NULL DEFAULT '', source_type TEXT NOT NULL DEFAULT 'manual', ca_provider TEXT NOT NULL DEFAULT 'letsencrypt', certificate TEXT NOT NULL DEFAULT '', private_key TEXT NOT NULL DEFAULT '', issuer TEXT NOT NULL DEFAULT '', not_after DATETIME, status TEXT NOT NULL DEFAULT 'pending', auto_renew BOOLEAN NOT NULL DEFAULT 1, renew_days_before INTEGER NOT NULL DEFAULT 30, last_error TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, acme_email TEXT NOT NULL DEFAULT '', key_algorithm TEXT NOT NULL DEFAULT 'rsa2048', cloud_account_id INTEGER, cloud_cert_id TEXT NOT NULL DEFAULT '', cert_chain TEXT NOT NULL DEFAULT '', not_before DATETIME, fingerprint TEXT NOT NULL DEFAULT '', dns_provider_id INTEGER, last_renew_at DATETIME);
CREATE TABLE ssl_dns_provider (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, provider TEXT NOT NULL, config TEXT NOT NULL DEFAULT '', email TEXT NOT NULL DEFAULT '', enabled BOOLEAN NOT NULL DEFAULT 1, last_test_at DATETIME, last_test_ok BOOLEAN NOT NULL DEFAULT 0, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME, phone TEXT NOT NULL DEFAULT '');
CREATE TABLE audit_operation_log (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '', module TEXT NOT NULL DEFAULT '', action TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '', method TEXT NOT NULL DEFAULT 'POST', path TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL DEFAULT 200, error_message TEXT NOT NULL DEFAULT '', cost_time INTEGER NOT NULL DEFAULT 0, ip TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE audit_login_log (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '', login_type TEXT NOT NULL DEFAULT 'web', status TEXT NOT NULL DEFAULT 'success', ip TEXT NOT NULL DEFAULT '', message TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE audit_data_log (id INTEGER PRIMARY KEY AUTOINCREMENT, user_id INTEGER NOT NULL DEFAULT 0, username TEXT NOT NULL DEFAULT '', table_name TEXT NOT NULL DEFAULT '', record_id INTEGER NOT NULL DEFAULT 0, action TEXT NOT NULL DEFAULT '', old_data TEXT NOT NULL DEFAULT '', new_data TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE kubernetes_cluster (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT NOT NULL, alias TEXT NOT NULL DEFAULT '', api_endpoint TEXT NOT NULL DEFAULT '', kube_config TEXT NOT NULL DEFAULT '', version TEXT NOT NULL DEFAULT '', status INTEGER NOT NULL DEFAULT 0, node_count INTEGER NOT NULL DEFAULT 0, pod_count INTEGER NOT NULL DEFAULT 0, region TEXT NOT NULL DEFAULT '', provider TEXT NOT NULL DEFAULT '', description TEXT NOT NULL DEFAULT '', created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, deleted_at DATETIME);
CREATE TABLE kubernetes_terminal_session (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, namespace TEXT NOT NULL DEFAULT '', pod_name TEXT NOT NULL DEFAULT '', container_name TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL, recording TEXT NOT NULL DEFAULT '', duration INTEGER NOT NULL DEFAULT 0, command_count INTEGER NOT NULL DEFAULT 0, status TEXT NOT NULL DEFAULT 'completed', started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE kubernetes_cluster_inspection (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, cluster_name TEXT NOT NULL DEFAULT '', status TEXT NOT NULL DEFAULT 'running', score INTEGER NOT NULL DEFAULT 0, check_count INTEGER NOT NULL DEFAULT 0, pass_count INTEGER NOT NULL DEFAULT 0, warning_count INTEGER NOT NULL DEFAULT 0, fail_count INTEGER NOT NULL DEFAULT 0, duration INTEGER NOT NULL DEFAULT 0, report_data TEXT NOT NULL DEFAULT '', user_id INTEGER NOT NULL, started_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, ended_at DATETIME, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP);
CREATE TABLE kubernetes_user_kube_config (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL, service_account TEXT NOT NULL, namespace TEXT NOT NULL DEFAULT 'gops-auth', is_active INTEGER NOT NULL DEFAULT 1, created_by INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, revoked_at DATETIME, UNIQUE(cluster_id, user_id, service_account));
CREATE TABLE kubernetes_user_role_binding (id INTEGER PRIMARY KEY AUTOINCREMENT, cluster_id INTEGER NOT NULL, user_id INTEGER NOT NULL, role_name TEXT NOT NULL, role_namespace TEXT NOT NULL DEFAULT '', role_type TEXT NOT NULL, bound_by INTEGER NOT NULL, created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, UNIQUE(cluster_id, user_id, role_name, role_namespace));
CREATE TABLE user_mfa (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 多因素认证配置ID */
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE, /* 用户ID */
  mfa_type TEXT NOT NULL, /* 认证类型：email、totp、passkey */
  name TEXT NOT NULL DEFAULT '', /* 认证器或凭据名称 */
  secret TEXT NOT NULL DEFAULT '', /* TOTP加密密钥 */
  credential_id TEXT NOT NULL DEFAULT '', /* WebAuthn凭据ID */
  credential_json TEXT NOT NULL DEFAULT '', /* WebAuthn凭据JSON */
  enabled INTEGER NOT NULL DEFAULT 0, /* 是否启用二次验证 */
  direct_login_enabled INTEGER NOT NULL DEFAULT 0, /* 是否允许直接登录 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  CHECK (mfa_type IN ('email', 'totp', 'passkey'))
);
CREATE TABLE ssl_deploy_config (
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
CREATE TABLE ssl_task (
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
CREATE TABLE "role" (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 角色ID */
  code VARCHAR(64) NOT NULL, /* 角色编码 */
  name VARCHAR(64) NOT NULL, /* 角色名称 */
  description VARCHAR(255), /* 角色说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE "user" (
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
CREATE TABLE casbin_rule (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* Casbin规则ID */
  ptype TEXT NOT NULL DEFAULT '', /* 策略类型 */
  v0 TEXT NOT NULL DEFAULT '', /* 策略字段0 */
  v1 TEXT NOT NULL DEFAULT '', /* 策略字段1 */
  v2 TEXT NOT NULL DEFAULT '', /* 策略字段2 */
  v3 TEXT NOT NULL DEFAULT '', /* 策略字段3 */
  v4 TEXT NOT NULL DEFAULT '', /* 策略字段4 */
  v5 TEXT NOT NULL DEFAULT '' /* 策略字段5 */
);
CREATE TABLE "monitor_domain" (
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
  response_threshold INTEGER NOT NULL DEFAULT 1000, /* 响应耗时告警阈值 */
  ssl_expiry_days INTEGER NOT NULL DEFAULT 30, /* SSL到期预警天数 */
  last_check DATETIME, /* 上次检查时间 */
  next_check DATETIME, /* 下次检查时间 */
  last_error TEXT NOT NULL DEFAULT '', /* 最近错误 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE "task_template" (
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
CREATE TABLE iot_topic (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* Topic ID */
  topic TEXT NOT NULL, /* Topic名称 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE iot_received_message (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 消息ID */
  received_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 接收时间 */
  client_id TEXT NOT NULL DEFAULT '', /* 客户端ID */
  username VARCHAR(255) NOT NULL DEFAULT '', /* MQTT用户名 */
  topic TEXT NOT NULL, /* Topic */
  payload TEXT NOT NULL DEFAULT '', /* 消息内容 */
  qos INTEGER NOT NULL DEFAULT 0, /* QoS */
  retained BOOLEAN NOT NULL DEFAULT FALSE, /* 是否保留 */
  duplicate BOOLEAN NOT NULL DEFAULT FALSE, /* 是否重复 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE iot_client_connection (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 客户端记录ID */
  client_id TEXT NOT NULL, /* MQTT客户端ID */
  status VARCHAR(16) NOT NULL DEFAULT 'disconnected', /* 连接状态 */
  username VARCHAR(255) NOT NULL DEFAULT '', /* MQTT用户名 */
  remote_address TEXT NOT NULL DEFAULT '', /* 远端地址 */
  listener VARCHAR(255) NOT NULL DEFAULT '', /* 监听器 */
  protocol_version INTEGER NOT NULL DEFAULT 0, /* MQTT协议版本 */
  clean_session BOOLEAN NOT NULL DEFAULT FALSE, /* 是否清理会话 */
  connected_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 连接时间 */
  disconnected_at DATETIME, /* 断开时间 */
  last_disconnected_ip TEXT NOT NULL DEFAULT '', /* 最后断开IP */
  disconnect_reason TEXT NOT NULL DEFAULT '', /* 断开原因 */
  session_expired BOOLEAN NOT NULL DEFAULT FALSE, /* 会话是否过期 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE iot_acl_rule (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* ACL规则ID */
  role_id INTEGER NOT NULL REFERENCES role(id) ON DELETE CASCADE, /* 系统角色ID */
  operation VARCHAR(16) NOT NULL, /* 操作：publish/subscribe */
  effect VARCHAR(8) NOT NULL, /* 效果：allow/deny */
  topic_filter TEXT NOT NULL, /* Topic过滤器 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE TABLE iot_external_broker (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 外部 Broker ID */
  name TEXT NOT NULL, /* Broker 名称 */
  url TEXT NOT NULL, /* Broker 连接地址 */
  client_id TEXT NOT NULL, /* MQTT Client ID */
  username TEXT NOT NULL DEFAULT '', /* MQTT 用户名 */
  password TEXT NOT NULL DEFAULT '', /* 加密后的 MQTT 密码 */
  topic_filter TEXT NOT NULL DEFAULT '#', /* 订阅 Topic 过滤器 */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用 */
  insecure_skip_verify INTEGER NOT NULL DEFAULT 0, /* 是否跳过 TLS 证书校验 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE INDEX idx_user_role_role_id ON "user_role" ("role_id");
CREATE UNIQUE INDEX idx_organization_code_active ON "organization" ("code") WHERE "deleted_at" IS NULL;
CREATE INDEX idx_organization_parent_id ON "organization" ("parent_id");
CREATE INDEX idx_organization_user_user_id ON "organization_user" ("user_id");
CREATE UNIQUE INDEX idx_asset_group_code_active ON asset_group (code) WHERE deleted_at IS NULL;
CREATE INDEX idx_host_group_active ON host (group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_monitor_check_history_domain_at ON monitor_check_history(domain_id, checked_at DESC);
CREATE INDEX idx_monitor_alert_log_sent_at ON monitor_alert_log(sent_at DESC);
CREATE INDEX idx_nginx_access_log_source_time ON nginx_access_log(source_id, occurred_at DESC);
CREATE INDEX idx_task_job_status ON task_job(status);
CREATE INDEX idx_task_execution_job ON task_execution(job_id, created_at DESC);
CREATE INDEX idx_ssl_certificate_domain ON ssl_certificate(domain);
CREATE INDEX idx_kubernetes_terminal_session_cluster ON kubernetes_terminal_session(cluster_id);
CREATE INDEX idx_kubernetes_terminal_session_user ON kubernetes_terminal_session(user_id);
CREATE INDEX idx_kubernetes_cluster_inspection_cluster ON kubernetes_cluster_inspection(cluster_id);
CREATE INDEX idx_kubernetes_user_kube_config_cluster ON kubernetes_user_kube_config(cluster_id);
CREATE INDEX idx_kubernetes_user_kube_config_user ON kubernetes_user_kube_config(user_id);
CREATE INDEX idx_kubernetes_user_role_binding_cluster ON kubernetes_user_role_binding(cluster_id);
CREATE INDEX idx_kubernetes_user_role_binding_user ON kubernetes_user_role_binding(user_id);
CREATE INDEX idx_db_instance_group_active ON db_instance (group_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_user_mfa_user_type ON user_mfa(user_id, mfa_type);
CREATE UNIQUE INDEX idx_user_mfa_singleton ON user_mfa(user_id, mfa_type) WHERE mfa_type IN ('email', 'totp');
CREATE UNIQUE INDEX idx_user_mfa_credential_id ON user_mfa(credential_id) WHERE credential_id <> '';
CREATE INDEX idx_ssl_deploy_config_certificate ON ssl_deploy_config(certificate_id);
CREATE INDEX idx_ssl_task_certificate ON ssl_task(certificate_id);
CREATE INDEX idx_ssl_task_status ON ssl_task(status);
CREATE UNIQUE INDEX idx_role_code_active ON role(code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_user_username_active ON user(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_user_email_active ON user(email) WHERE deleted_at IS NULL AND email <> '';
CREATE UNIQUE INDEX idx_casbin_rule ON casbin_rule(ptype, v0, v1, v2, v3, v4, v5);
CREATE INDEX idx_monitor_domain_next_check ON monitor_domain(next_check);
CREATE UNIQUE INDEX idx_monitor_domain_address_active ON monitor_domain(domain, port) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_task_template_code_active ON task_template(code) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_iot_topic_active ON iot_topic(topic) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_received_message_received_at ON iot_received_message(received_at DESC);
CREATE INDEX idx_iot_received_message_topic ON iot_received_message(topic);
CREATE INDEX idx_iot_received_message_client_id ON iot_received_message(client_id);
CREATE UNIQUE INDEX idx_iot_client_connection_client_active ON iot_client_connection(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_client_connection_connected_at ON iot_client_connection(connected_at DESC);
CREATE UNIQUE INDEX idx_iot_acl_rule_active ON iot_acl_rule(role_id, operation, effect, topic_filter) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_acl_rule_role_id ON iot_acl_rule(role_id);
CREATE UNIQUE INDEX idx_iot_external_broker_name_active ON iot_external_broker(name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_iot_external_broker_client_id_active ON iot_external_broker(client_id) WHERE deleted_at IS NULL;

INSERT INTO organization(name, code, manager, status)
SELECT '默认组织', 'root', '', TRUE
WHERE NOT EXISTS (SELECT 1 FROM organization WHERE code = 'root' AND deleted_at IS NULL);

INSERT INTO system_config(config_key, config_value) VALUES ('system_name', 'Gops') ON CONFLICT(config_key) DO NOTHING;
INSERT INTO system_config(config_key, config_value) VALUES ('system_description', '') ON CONFLICT(config_key) DO NOTHING;

-- 初始化超级管理员，初始密码为 password；首次登录后应立即修改。
INSERT INTO role(code, name, description)
SELECT 'admin', '超级管理员', '拥有全部后台接口权限'
WHERE NOT EXISTS (SELECT 1 FROM role WHERE code = 'admin' AND deleted_at IS NULL);

INSERT INTO user(username, password_hash, nickname, email, status)
SELECT 'admin', '$2a$10$itQzLKzPEE0KGCYXAiKol.rOt0GjhsQkm6mX42SOk09Hhye5DjM5u', '管理员', '', TRUE
WHERE NOT EXISTS (SELECT 1 FROM user WHERE username = 'admin' AND deleted_at IS NULL);

INSERT INTO user_role(user_id, role_id)
SELECT users.id, roles.id
FROM user AS users
JOIN role AS roles ON roles.code = 'admin' AND roles.deleted_at IS NULL
WHERE users.username = 'admin' AND users.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM user_role AS relation
    WHERE relation.user_id = users.id AND relation.role_id = roles.id
  );

INSERT INTO organization_user(organization_id, user_id)
SELECT organizations.id, users.id
FROM organization AS organizations
JOIN user AS users ON users.username = 'admin' AND users.deleted_at IS NULL
WHERE organizations.code = 'root' AND organizations.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM organization_user AS relation
    WHERE relation.organization_id = organizations.id AND relation.user_id = users.id
  );

INSERT INTO casbin_rule(ptype, v0, v1, v2, v3, v4, v5)
SELECT 'g', CAST(users.id AS TEXT), 'admin', '', '', '', ''
FROM user AS users
WHERE users.username = 'admin' AND users.deleted_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM casbin_rule
    WHERE ptype = 'g' AND v0 = CAST(users.id AS TEXT) AND v1 = 'admin'
  );

INSERT INTO casbin_rule(ptype, v0, v1, v2, v3, v4, v5)
SELECT 'p', 'admin', '/api/v1/*', '.*', '', '', ''
WHERE NOT EXISTS (
  SELECT 1 FROM casbin_rule
  WHERE ptype = 'p' AND v0 = 'admin' AND v1 = '/api/v1/*' AND v2 = '.*'
);

-- +goose Down
PRAGMA foreign_keys = OFF;
DROP TABLE IF EXISTS "casbin_rule";
DROP TABLE IF EXISTS "iot_external_broker";
DROP TABLE IF EXISTS "iot_acl_rule";
DROP TABLE IF EXISTS "iot_client_connection";
DROP TABLE IF EXISTS "iot_received_message";
DROP TABLE IF EXISTS "iot_topic";
DROP TABLE IF EXISTS "task_template";
DROP TABLE IF EXISTS "monitor_domain";
DROP TABLE IF EXISTS "user";
DROP TABLE IF EXISTS "role";
DROP TABLE IF EXISTS "ssl_task";
DROP TABLE IF EXISTS "ssl_deploy_config";
DROP TABLE IF EXISTS "user_mfa";
DROP TABLE IF EXISTS "kubernetes_user_role_binding";
DROP TABLE IF EXISTS "kubernetes_user_kube_config";
DROP TABLE IF EXISTS "kubernetes_cluster_inspection";
DROP TABLE IF EXISTS "kubernetes_terminal_session";
DROP TABLE IF EXISTS "kubernetes_cluster";
DROP TABLE IF EXISTS "audit_data_log";
DROP TABLE IF EXISTS "audit_login_log";
DROP TABLE IF EXISTS "audit_operation_log";
DROP TABLE IF EXISTS "ssl_dns_provider";
DROP TABLE IF EXISTS "ssl_certificate";
DROP TABLE IF EXISTS "task_execution";
DROP TABLE IF EXISTS "task_job";
DROP TABLE IF EXISTS "nginx_access_log";
DROP TABLE IF EXISTS "nginx_source";
DROP TABLE IF EXISTS "monitor_alert_log";
DROP TABLE IF EXISTS "monitor_alert_receiver";
DROP TABLE IF EXISTS "monitor_alert_channel";
DROP TABLE IF EXISTS "monitor_check_history";
DROP TABLE IF EXISTS "terminal_session";
DROP TABLE IF EXISTS "db_instance";
DROP TABLE IF EXISTS "host";
DROP TABLE IF EXISTS "cloud_account";
DROP TABLE IF EXISTS "credential";
DROP TABLE IF EXISTS "asset_group";
DROP TABLE IF EXISTS "system_config";
DROP TABLE IF EXISTS "organization_user";
DROP TABLE IF EXISTS "organization";
DROP TABLE IF EXISTS "user_role";
PRAGMA foreign_keys = ON;
