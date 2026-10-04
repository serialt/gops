-- +goose Up
CREATE TABLE iot_role (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* IOT角色ID */
  name VARCHAR(255) NOT NULL, /* 角色名称 */
  description VARCHAR(500) NOT NULL DEFAULT '', /* 角色说明 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX idx_iot_role_name_active ON iot_role(name) WHERE deleted_at IS NULL;

CREATE TABLE iot_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* IOT用户ID */
  username VARCHAR(255) NOT NULL, /* MQTT用户名 */
  password_hash VARCHAR(255) NOT NULL, /* 密码哈希 */
  enabled BOOLEAN NOT NULL DEFAULT TRUE, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX idx_iot_user_username_active ON iot_user(username) WHERE deleted_at IS NULL;

CREATE TABLE iot_user_role (
  user_id INTEGER NOT NULL REFERENCES iot_user(id) ON DELETE CASCADE, /* IOT用户ID */
  role_id INTEGER NOT NULL REFERENCES iot_role(id) ON DELETE CASCADE, /* IOT角色ID */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_iot_user_role_role_id ON iot_user_role(role_id);

CREATE TABLE iot_acl_rule (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* ACL规则ID */
  role_id INTEGER NOT NULL REFERENCES iot_role(id) ON DELETE CASCADE, /* IOT角色ID */
  operation VARCHAR(16) NOT NULL, /* 操作：publish/subscribe */
  effect VARCHAR(8) NOT NULL, /* 效果：allow/deny */
  topic_filter TEXT NOT NULL, /* Topic过滤器 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX idx_iot_acl_rule_active ON iot_acl_rule(role_id, operation, effect, topic_filter) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_acl_rule_role_id ON iot_acl_rule(role_id);

CREATE TABLE iot_topic (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* Topic ID */
  topic TEXT NOT NULL, /* Topic名称 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX idx_iot_topic_active ON iot_topic(topic) WHERE deleted_at IS NULL;

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
CREATE INDEX idx_iot_received_message_received_at ON iot_received_message(received_at DESC);
CREATE INDEX idx_iot_received_message_topic ON iot_received_message(topic);
CREATE INDEX idx_iot_received_message_client_id ON iot_received_message(client_id);

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
CREATE UNIQUE INDEX idx_iot_client_connection_client_active ON iot_client_connection(client_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_client_connection_connected_at ON iot_client_connection(connected_at DESC);

-- +goose Down
DROP TABLE IF EXISTS iot_client_connection;
DROP TABLE IF EXISTS iot_received_message;
DROP TABLE IF EXISTS iot_topic;
DROP TABLE IF EXISTS iot_acl_rule;
DROP TABLE IF EXISTS iot_user_role;
DROP TABLE IF EXISTS iot_user;
DROP TABLE IF EXISTS iot_role;
