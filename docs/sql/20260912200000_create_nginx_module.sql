-- +goose Up
-- Nginx 数据源与访问日志模块。
CREATE TABLE IF NOT EXISTS nginx_source (
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
CREATE TABLE IF NOT EXISTS nginx_access_log (
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
CREATE INDEX IF NOT EXISTS idx_nginx_access_log_source_time ON nginx_access_log(source_id, occurred_at DESC);
-- +goose Down
DROP TABLE IF EXISTS nginx_access_log;
DROP TABLE IF EXISTS nginx_source;
