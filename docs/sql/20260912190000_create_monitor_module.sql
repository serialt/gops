-- +goose Up
-- 表：域名监控表
CREATE TABLE IF NOT EXISTS monitor_domain (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 监控主键
  domain VARCHAR(255) NOT NULL, -- 域名或地址
  port INTEGER NOT NULL DEFAULT 443, -- 服务端口
  status VARCHAR(20) NOT NULL DEFAULT 'unknown', -- 监控状态：normal、abnormal、paused、unknown
  response_time INTEGER NOT NULL DEFAULT 0, -- 最近响应耗时（毫秒）
  ssl_valid BOOLEAN NOT NULL DEFAULT FALSE, -- SSL 证书是否有效
  ssl_expiry DATETIME, -- SSL 证书到期时间
  check_interval INTEGER NOT NULL DEFAULT 300, -- 检查间隔（秒）
  enable_ssl BOOLEAN NOT NULL DEFAULT TRUE, -- 是否检查 SSL 证书
  enable_alert BOOLEAN NOT NULL DEFAULT FALSE, -- 是否启用告警
  response_threshold INTEGER NOT NULL DEFAULT 1000, -- 响应耗时告警阈值（毫秒）
  ssl_expiry_days INTEGER NOT NULL DEFAULT 30, -- SSL 到期告警提前天数
  last_check DATETIME, -- 最后检查时间
  next_check DATETIME, -- 下次检查时间
  last_error TEXT NOT NULL DEFAULT '', -- 最近检查错误信息
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME, -- 软删除时间
  UNIQUE(domain, port)
);
CREATE INDEX IF NOT EXISTS idx_monitor_domain_next_check ON monitor_domain(next_check);

-- 表：监控检查历史表
CREATE TABLE IF NOT EXISTS monitor_check_history (
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
CREATE INDEX IF NOT EXISTS idx_monitor_check_history_domain_at ON monitor_check_history(domain_id, checked_at DESC);

-- 表：告警通道表
CREATE TABLE IF NOT EXISTS monitor_alert_channel (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 通道主键
  name VARCHAR(100) NOT NULL, -- 通道名称
  channel_type VARCHAR(20) NOT NULL, -- 通道类型：email、webhook、wechat、dingtalk、feishu
  enabled BOOLEAN NOT NULL DEFAULT TRUE, -- 是否启用
  config TEXT NOT NULL DEFAULT '{}', -- 通道 JSON 配置
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);

-- 表：告警接收人表
CREATE TABLE IF NOT EXISTS monitor_alert_receiver (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 接收人主键
  name VARCHAR(100) NOT NULL, -- 接收人名称
  email VARCHAR(255) NOT NULL DEFAULT '', -- 邮箱地址
  phone VARCHAR(30) NOT NULL DEFAULT '', -- 手机号码
  webhook VARCHAR(500) NOT NULL DEFAULT '', -- 个人 Webhook 地址
  enabled BOOLEAN NOT NULL DEFAULT TRUE, -- 是否启用
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME -- 软删除时间
);

-- 表：告警日志表
CREATE TABLE IF NOT EXISTS monitor_alert_log (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 告警日志主键
  domain_id INTEGER REFERENCES monitor_domain(id) ON DELETE SET NULL, -- 域名监控主键
  alert_type VARCHAR(50) NOT NULL, -- 告警类型
  status VARCHAR(20) NOT NULL, -- 发送状态：success、failed
  message TEXT NOT NULL DEFAULT '', -- 告警内容
  channel_type VARCHAR(20) NOT NULL DEFAULT '', -- 发送通道类型
  error_message TEXT NOT NULL DEFAULT '', -- 发送失败原因
  sent_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP -- 发送时间
);
CREATE INDEX IF NOT EXISTS idx_monitor_alert_log_sent_at ON monitor_alert_log(sent_at DESC);

-- +goose Down
DROP TABLE IF EXISTS monitor_alert_log;
DROP TABLE IF EXISTS monitor_alert_receiver;
DROP TABLE IF EXISTS monitor_alert_channel;
DROP TABLE IF EXISTS monitor_check_history;
DROP TABLE IF EXISTS monitor_domain;
