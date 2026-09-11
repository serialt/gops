-- +goose Up
CREATE TABLE IF NOT EXISTS asset_group (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  code VARCHAR(50) NOT NULL,
  parent_id BIGINT NOT NULL DEFAULT 0,
  description VARCHAR(500) NOT NULL DEFAULT '',
  sort_order INTEGER NOT NULL DEFAULT 0,
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_asset_group_code_active ON asset_group (code) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS credential (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  credential_type VARCHAR(20) NOT NULL,
  username VARCHAR(100) NOT NULL DEFAULT '',
  password TEXT NOT NULL DEFAULT '',
  private_key TEXT NOT NULL DEFAULT '',
  passphrase TEXT NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS cloud_account (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  provider VARCHAR(50) NOT NULL,
  access_key TEXT NOT NULL DEFAULT '',
  secret_key TEXT NOT NULL DEFAULT '',
  region VARCHAR(100) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ
);

CREATE TABLE IF NOT EXISTS host (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  group_id BIGINT,
  host_type VARCHAR(20) NOT NULL DEFAULT 'self',
  ip VARCHAR(255) NOT NULL,
  port INTEGER NOT NULL DEFAULT 22,
  ssh_user VARCHAR(100) NOT NULL DEFAULT '',
  credential_id BIGINT,
  cloud_account_id BIGINT,
  tags VARCHAR(500) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status SMALLINT NOT NULL DEFAULT -1,
  last_seen_at TIMESTAMPTZ,
  os VARCHAR(100) NOT NULL DEFAULT '',
  kernel VARCHAR(100) NOT NULL DEFAULT '',
  arch VARCHAR(50) NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_host_group FOREIGN KEY (group_id) REFERENCES asset_group(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL,
  CONSTRAINT fk_host_cloud_account FOREIGN KEY (cloud_account_id) REFERENCES cloud_account(id) ON DELETE SET NULL
);
CREATE INDEX IF NOT EXISTS idx_host_group_active ON host (group_id) WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS db_instance (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  name VARCHAR(255) NOT NULL,
  db_type VARCHAR(50) NOT NULL,
  host VARCHAR(255) NOT NULL,
  port INTEGER NOT NULL,
  username VARCHAR(255) NOT NULL DEFAULT '',
  password TEXT NOT NULL DEFAULT '',
  version VARCHAR(50) NOT NULL DEFAULT '',
  credential_id BIGINT,
  tags VARCHAR(255) NOT NULL DEFAULT '',
  description VARCHAR(500) NOT NULL DEFAULT '',
  status BOOLEAN NOT NULL DEFAULT TRUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_db_instance_credential FOREIGN KEY (credential_id) REFERENCES credential(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS terminal_session (
  id BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  user_id BIGINT NOT NULL,
  host_id BIGINT,
  db_instance_id BIGINT,
  session_type VARCHAR(20) NOT NULL,
  started_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  ended_at TIMESTAMPTZ,
  duration_seconds BIGINT NOT NULL DEFAULT 0,
  command_count BIGINT NOT NULL DEFAULT 0,
  recording TEXT NOT NULL DEFAULT '',
  created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at TIMESTAMPTZ,
  CONSTRAINT fk_terminal_session_host FOREIGN KEY (host_id) REFERENCES host(id) ON DELETE SET NULL,
  CONSTRAINT fk_terminal_session_db FOREIGN KEY (db_instance_id) REFERENCES db_instance(id) ON DELETE SET NULL
);

COMMENT ON TABLE asset_group IS '资产分组表';
COMMENT ON TABLE credential IS '资产凭证表';
COMMENT ON TABLE cloud_account IS '云账号表';
COMMENT ON TABLE host IS '主机资产表';
COMMENT ON TABLE db_instance IS '数据库资产表';
COMMENT ON TABLE terminal_session IS '终端审计会话表';
COMMENT ON COLUMN asset_group.name IS '分组名称';
COMMENT ON COLUMN asset_group.code IS '分组编码';
COMMENT ON COLUMN asset_group.parent_id IS '父分组ID';
COMMENT ON COLUMN asset_group.description IS '分组描述';
COMMENT ON COLUMN asset_group.sort_order IS '排序值';
COMMENT ON COLUMN asset_group.status IS '是否启用';
COMMENT ON COLUMN asset_group.created_at IS '创建时间';
COMMENT ON COLUMN asset_group.updated_at IS '更新时间';
COMMENT ON COLUMN asset_group.deleted_at IS '软删除时间';
COMMENT ON COLUMN credential.name IS '凭证名称';
COMMENT ON COLUMN credential.credential_type IS '凭证类型';
COMMENT ON COLUMN credential.username IS '用户名';
COMMENT ON COLUMN credential.password IS '加密密码';
COMMENT ON COLUMN credential.private_key IS '加密私钥';
COMMENT ON COLUMN credential.passphrase IS '加密私钥口令';
COMMENT ON COLUMN credential.description IS '凭证描述';
COMMENT ON COLUMN credential.created_at IS '创建时间';
COMMENT ON COLUMN credential.updated_at IS '更新时间';
COMMENT ON COLUMN credential.deleted_at IS '软删除时间';
COMMENT ON COLUMN cloud_account.name IS '账号名称';
COMMENT ON COLUMN cloud_account.provider IS '云服务商';
COMMENT ON COLUMN cloud_account.access_key IS '加密访问密钥';
COMMENT ON COLUMN cloud_account.secret_key IS '加密访问密钥密码';
COMMENT ON COLUMN cloud_account.region IS '默认区域';
COMMENT ON COLUMN cloud_account.description IS '账号描述';
COMMENT ON COLUMN cloud_account.status IS '是否启用';
COMMENT ON COLUMN cloud_account.created_at IS '创建时间';
COMMENT ON COLUMN cloud_account.updated_at IS '更新时间';
COMMENT ON COLUMN cloud_account.deleted_at IS '软删除时间';
COMMENT ON COLUMN host.name IS '主机名称';
COMMENT ON COLUMN host.group_id IS '所属资产分组ID';
COMMENT ON COLUMN host.host_type IS '主机类型';
COMMENT ON COLUMN host.ip IS '主机地址';
COMMENT ON COLUMN host.port IS 'SSH端口';
COMMENT ON COLUMN host.ssh_user IS 'SSH用户名';
COMMENT ON COLUMN host.credential_id IS '关联凭证ID';
COMMENT ON COLUMN host.cloud_account_id IS '关联云账号ID';
COMMENT ON COLUMN host.tags IS '主机标签';
COMMENT ON COLUMN host.description IS '主机描述';
COMMENT ON COLUMN host.status IS '连接状态';
COMMENT ON COLUMN host.last_seen_at IS '最后连接时间';
COMMENT ON COLUMN host.os IS '操作系统';
COMMENT ON COLUMN host.kernel IS '内核版本';
COMMENT ON COLUMN host.arch IS '系统架构';
COMMENT ON COLUMN host.created_at IS '创建时间';
COMMENT ON COLUMN host.updated_at IS '更新时间';
COMMENT ON COLUMN host.deleted_at IS '软删除时间';
COMMENT ON COLUMN db_instance.name IS '数据库名称';
COMMENT ON COLUMN db_instance.db_type IS '数据库类型';
COMMENT ON COLUMN db_instance.host IS '数据库地址';
COMMENT ON COLUMN db_instance.port IS '数据库端口';
COMMENT ON COLUMN db_instance.username IS '数据库用户名';
COMMENT ON COLUMN db_instance.password IS '加密数据库密码';
COMMENT ON COLUMN db_instance.version IS '数据库版本';
COMMENT ON COLUMN db_instance.credential_id IS '关联凭证ID';
COMMENT ON COLUMN db_instance.tags IS '数据库标签';
COMMENT ON COLUMN db_instance.description IS '数据库描述';
COMMENT ON COLUMN db_instance.status IS '是否启用';
COMMENT ON COLUMN db_instance.created_at IS '创建时间';
COMMENT ON COLUMN db_instance.updated_at IS '更新时间';
COMMENT ON COLUMN db_instance.deleted_at IS '软删除时间';
COMMENT ON COLUMN terminal_session.user_id IS '操作用户ID';
COMMENT ON COLUMN terminal_session.host_id IS '关联主机ID';
COMMENT ON COLUMN terminal_session.db_instance_id IS '关联数据库ID';
COMMENT ON COLUMN terminal_session.session_type IS '终端类型';
COMMENT ON COLUMN terminal_session.started_at IS '开始时间';
COMMENT ON COLUMN terminal_session.ended_at IS '结束时间';
COMMENT ON COLUMN terminal_session.duration_seconds IS '会话时长秒数';
COMMENT ON COLUMN terminal_session.command_count IS '命令数量';
COMMENT ON COLUMN terminal_session.recording IS '会话录制内容';
COMMENT ON COLUMN terminal_session.created_at IS '创建时间';
COMMENT ON COLUMN terminal_session.updated_at IS '更新时间';
COMMENT ON COLUMN terminal_session.deleted_at IS '软删除时间';

-- +goose Down
DROP TABLE IF EXISTS terminal_session;
DROP TABLE IF EXISTS db_instance;
DROP TABLE IF EXISTS host;
DROP TABLE IF EXISTS cloud_account;
DROP TABLE IF EXISTS credential;
DROP TABLE IF EXISTS asset_group;
