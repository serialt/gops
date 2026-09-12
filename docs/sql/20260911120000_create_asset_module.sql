-- +goose Up
-- 表：资产分组表
CREATE TABLE IF NOT EXISTS asset_group (
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
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_group_code_active ON asset_group (code) WHERE deleted_at IS NULL;

-- 表：资产凭证表
CREATE TABLE IF NOT EXISTS credential (
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

-- 表：云账号表
CREATE TABLE IF NOT EXISTS cloud_account (
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

-- 表：主机资产表
CREATE TABLE IF NOT EXISTS host (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 主机主键
  name VARCHAR(100) NOT NULL, -- 主机名称
  group_id BIGINT, -- 所属资产分组主键
  host_type VARCHAR(20) NOT NULL DEFAULT 'self', -- 主机类型
  ip VARCHAR(255) NOT NULL, -- 主机地址
  port INTEGER NOT NULL DEFAULT 22, -- SSH 端口
  ssh_user VARCHAR(100) NOT NULL DEFAULT '', -- SSH 用户名
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
  deleted_at DATETIME, -- 软删除时间
  CONSTRAINT fk_host_group FOREIGN KEY (group_id) REFERENCES asset_group(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_cloud_account FOREIGN KEY (cloud_account_id) REFERENCES cloud_account(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_host_group_active ON host (group_id) WHERE deleted_at IS NULL;

-- 表：数据库资产表
CREATE TABLE IF NOT EXISTS db_instance (
  id INTEGER PRIMARY KEY AUTOINCREMENT, -- 数据库主键
  name VARCHAR(255) NOT NULL, -- 数据库名称
  db_type VARCHAR(50) NOT NULL, -- 数据库类型
  host VARCHAR(255) NOT NULL, -- 数据库地址
  port INTEGER NOT NULL, -- 数据库端口
  username VARCHAR(255) NOT NULL DEFAULT '', -- 数据库用户名
  password TEXT NOT NULL DEFAULT '', -- 加密数据库密码
  version VARCHAR(50) NOT NULL DEFAULT '', -- 数据库版本
  credential_id BIGINT, -- 关联凭证主键
  tags VARCHAR(255) NOT NULL DEFAULT '', -- 数据库标签
  description VARCHAR(500) NOT NULL DEFAULT '', -- 数据库说明
  status BOOLEAN NOT NULL DEFAULT TRUE, -- 启用状态
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 创建时间
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 更新时间
  deleted_at DATETIME, -- 软删除时间
  CONSTRAINT fk_db_instance_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL
);

-- 表：终端审计会话表
CREATE TABLE IF NOT EXISTS terminal_session (
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


-- +goose Down
DROP TABLE IF EXISTS terminal_session;
DROP TABLE IF EXISTS db_instance;
DROP TABLE IF EXISTS host;
DROP TABLE IF EXISTS cloud_account;
DROP TABLE IF EXISTS credential;
DROP TABLE IF EXISTS asset_group;
