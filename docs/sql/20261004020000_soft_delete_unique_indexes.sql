-- +goose NO TRANSACTION
-- +goose Up
-- SQLite 的复合 UNIQUE(..., deleted_at) 无法约束 deleted_at 为 NULL 的活动记录，
-- 因此软删除表统一使用 WHERE deleted_at IS NULL 的部分唯一索引。
PRAGMA foreign_keys = OFF;

CREATE TABLE role_unique_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 角色ID */
  code VARCHAR(64) NOT NULL, /* 角色编码 */
  name VARCHAR(64) NOT NULL, /* 角色名称 */
  description VARCHAR(255), /* 角色说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
INSERT INTO role_unique_new(id, code, name, description, created_at, updated_at, deleted_at)
SELECT id, code, name, description, created_at, updated_at, deleted_at FROM role;
DROP TABLE role;
ALTER TABLE role_unique_new RENAME TO role;
CREATE UNIQUE INDEX idx_role_code_active ON role(code) WHERE deleted_at IS NULL;

CREATE TABLE user_unique_new (
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
INSERT INTO user_unique_new(id, username, password_hash, nickname, email, status, created_at, updated_at, deleted_at)
SELECT id, username, password_hash, nickname, email, status, created_at, updated_at, deleted_at FROM user;
DROP TABLE user;
ALTER TABLE user_unique_new RENAME TO user;
CREATE UNIQUE INDEX idx_user_username_active ON user(username) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_user_email_active ON user(email) WHERE deleted_at IS NULL AND email <> '';

CREATE TABLE monitor_domain_unique_new (
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
INSERT INTO monitor_domain_unique_new(id, domain, port, request_client, status, response_time, ssl_valid, ssl_expiry, ssl_root_ca, check_interval, check_jitter, enable_ssl, verify_certificate, enable_alert, response_threshold, ssl_expiry_days, last_check, next_check, last_error, created_at, updated_at, deleted_at)
SELECT id, domain, port, request_client, status, response_time, ssl_valid, ssl_expiry, ssl_root_ca, check_interval, check_jitter, enable_ssl, verify_certificate, enable_alert, response_threshold, ssl_expiry_days, last_check, next_check, last_error, created_at, updated_at, deleted_at FROM monitor_domain;
DROP TABLE monitor_domain;
ALTER TABLE monitor_domain_unique_new RENAME TO monitor_domain;
CREATE INDEX idx_monitor_domain_next_check ON monitor_domain(next_check);
CREATE UNIQUE INDEX idx_monitor_domain_address_active ON monitor_domain(domain, port) WHERE deleted_at IS NULL;

CREATE TABLE task_template_unique_new (
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
INSERT INTO task_template_unique_new(id, name, code, description, content, variables, category, platform, timeout, sort_order, status, created_by, created_at, updated_at, deleted_at)
SELECT id, name, code, description, content, variables, category, platform, timeout, sort_order, status, created_by, created_at, updated_at, deleted_at FROM task_template;
DROP TABLE task_template;
ALTER TABLE task_template_unique_new RENAME TO task_template;
CREATE UNIQUE INDEX idx_task_template_code_active ON task_template(code) WHERE deleted_at IS NULL;

PRAGMA foreign_keys = ON;

-- +goose Down
-- 不可逆迁移：回退为表级 UNIQUE 会在存在同值软删除历史记录时丢失数据或失败。
