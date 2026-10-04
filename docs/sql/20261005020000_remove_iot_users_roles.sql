-- +goose Up
DROP INDEX IF EXISTS idx_iot_acl_rule_active;
DROP INDEX IF EXISTS idx_iot_acl_rule_role_id;
ALTER TABLE iot_acl_rule RENAME TO iot_acl_rule_legacy;

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
INSERT INTO iot_acl_rule(id, role_id, operation, effect, topic_filter, created_at, updated_at, deleted_at)
SELECT acl.id, system_role.id, acl.operation, acl.effect, acl.topic_filter, acl.created_at, acl.updated_at, acl.deleted_at
FROM iot_acl_rule_legacy acl
JOIN iot_role legacy_role ON legacy_role.id = acl.role_id
JOIN role system_role ON system_role.name = legacy_role.name AND system_role.deleted_at IS NULL;
DROP TABLE iot_acl_rule_legacy;
CREATE UNIQUE INDEX idx_iot_acl_rule_active ON iot_acl_rule(role_id, operation, effect, topic_filter) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_acl_rule_role_id ON iot_acl_rule(role_id);

DROP TABLE IF EXISTS iot_user_role;
DROP TABLE IF EXISTS iot_user;
DROP TABLE IF EXISTS iot_role;

-- +goose Down
CREATE TABLE iot_role (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  name VARCHAR(255) NOT NULL,
  description VARCHAR(500) NOT NULL DEFAULT '',
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX idx_iot_role_name_active ON iot_role(name) WHERE deleted_at IS NULL;
CREATE TABLE iot_user (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username VARCHAR(255) NOT NULL,
  password_hash VARCHAR(255) NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT TRUE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
CREATE UNIQUE INDEX idx_iot_user_username_active ON iot_user(username) WHERE deleted_at IS NULL;
CREATE TABLE iot_user_role (
  user_id INTEGER NOT NULL REFERENCES iot_user(id) ON DELETE CASCADE,
  role_id INTEGER NOT NULL REFERENCES iot_role(id) ON DELETE CASCADE,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (user_id, role_id)
);
CREATE INDEX idx_iot_user_role_role_id ON iot_user_role(role_id);

DROP INDEX IF EXISTS idx_iot_acl_rule_active;
DROP INDEX IF EXISTS idx_iot_acl_rule_role_id;
ALTER TABLE iot_acl_rule RENAME TO iot_acl_rule_global;
CREATE TABLE iot_acl_rule (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  role_id INTEGER NOT NULL REFERENCES iot_role(id) ON DELETE CASCADE,
  operation VARCHAR(16) NOT NULL,
  effect VARCHAR(8) NOT NULL,
  topic_filter TEXT NOT NULL,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  deleted_at DATETIME
);
INSERT INTO iot_role (id, name, description, created_at, updated_at)
SELECT DISTINCT role.id, role.name, role.description, role.created_at, role.updated_at
FROM role JOIN iot_acl_rule_global ON iot_acl_rule_global.role_id = role.id;
INSERT INTO iot_acl_rule (id, role_id, operation, effect, topic_filter, created_at, updated_at, deleted_at)
SELECT id, role_id, operation, effect, topic_filter, created_at, updated_at, deleted_at FROM iot_acl_rule_global;
DROP TABLE iot_acl_rule_global;
CREATE UNIQUE INDEX idx_iot_acl_rule_active ON iot_acl_rule(role_id, operation, effect, topic_filter) WHERE deleted_at IS NULL;
CREATE INDEX idx_iot_acl_rule_role_id ON iot_acl_rule(role_id);
