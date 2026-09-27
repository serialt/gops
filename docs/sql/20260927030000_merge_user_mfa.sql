-- +goose Up
-- 将邮箱、TOTP、Passkey 三类用户多因素认证配置合并到统一表。

CREATE TABLE IF NOT EXISTS user_mfa (
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
CREATE INDEX IF NOT EXISTS idx_user_mfa_user_type ON user_mfa(user_id, mfa_type);
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_mfa_singleton ON user_mfa(user_id, mfa_type) WHERE mfa_type IN ('email', 'totp');
CREATE UNIQUE INDEX IF NOT EXISTS idx_user_mfa_credential_id ON user_mfa(credential_id) WHERE credential_id <> '';

-- 新库会创建空的兼容表；已有数据库会复用旧表并迁移其中的数据。
CREATE TABLE IF NOT EXISTS user_mfa_totp (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 旧TOTP配置ID */
  user_id INTEGER NOT NULL UNIQUE, /* 用户ID */
  name TEXT NOT NULL DEFAULT '身份验证器', /* 验证器名称 */
  secret TEXT NOT NULL DEFAULT '', /* TOTP加密密钥 */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE TABLE IF NOT EXISTS user_mfa_passkey (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 旧Passkey配置ID */
  user_id INTEGER NOT NULL, /* 用户ID */
  name TEXT NOT NULL DEFAULT '', /* Passkey名称 */
  credential_id TEXT NOT NULL UNIQUE, /* WebAuthn凭据ID */
  credential_json TEXT NOT NULL, /* WebAuthn凭据JSON */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用二次验证 */
  direct_login_enabled INTEGER NOT NULL DEFAULT 0, /* 是否允许直接登录 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE TABLE IF NOT EXISTS user_mfa_email (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 旧邮箱MFA配置ID */
  user_id INTEGER NOT NULL UNIQUE, /* 用户ID */
  mfa_enabled INTEGER NOT NULL DEFAULT 0, /* 是否启用二次验证 */
  direct_login_enabled INTEGER NOT NULL DEFAULT 0, /* 是否允许直接登录 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);

INSERT OR IGNORE INTO user_mfa(user_id, mfa_type, name, secret, enabled, direct_login_enabled, created_at, updated_at)
SELECT user_id, 'totp', name, secret, enabled, 0, created_at, updated_at FROM user_mfa_totp;
INSERT OR IGNORE INTO user_mfa(user_id, mfa_type, name, credential_id, credential_json, enabled, direct_login_enabled, created_at, updated_at)
SELECT user_id, 'passkey', name, credential_id, credential_json, enabled, direct_login_enabled, created_at, updated_at FROM user_mfa_passkey;
INSERT OR IGNORE INTO user_mfa(user_id, mfa_type, name, enabled, direct_login_enabled, created_at, updated_at)
SELECT user_id, 'email', '邮箱验证码', mfa_enabled, direct_login_enabled, created_at, updated_at FROM user_mfa_email;

DROP TABLE IF EXISTS user_mfa_email;
DROP TABLE IF EXISTS user_mfa_passkey;
DROP TABLE IF EXISTS user_mfa_totp;

-- +goose Down
CREATE TABLE IF NOT EXISTS user_mfa_totp (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* TOTP配置ID */
  user_id INTEGER NOT NULL UNIQUE REFERENCES user(id) ON DELETE CASCADE, /* 用户ID */
  name TEXT NOT NULL DEFAULT '身份验证器', /* 验证器名称 */
  secret TEXT NOT NULL DEFAULT '', /* TOTP加密密钥 */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE TABLE IF NOT EXISTS user_mfa_passkey (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* Passkey配置ID */
  user_id INTEGER NOT NULL REFERENCES user(id) ON DELETE CASCADE, /* 用户ID */
  name TEXT NOT NULL DEFAULT '', /* Passkey名称 */
  credential_id TEXT NOT NULL UNIQUE, /* WebAuthn凭据ID */
  credential_json TEXT NOT NULL, /* WebAuthn凭据JSON */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用二次验证 */
  direct_login_enabled INTEGER NOT NULL DEFAULT 0, /* 是否允许直接登录 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);
CREATE TABLE IF NOT EXISTS user_mfa_email (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 邮箱MFA配置ID */
  user_id INTEGER NOT NULL UNIQUE REFERENCES user(id) ON DELETE CASCADE, /* 用户ID */
  mfa_enabled INTEGER NOT NULL DEFAULT 0, /* 是否启用二次验证 */
  direct_login_enabled INTEGER NOT NULL DEFAULT 0, /* 是否允许直接登录 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP /* 更新时间 */
);

INSERT OR IGNORE INTO user_mfa_totp(user_id, name, secret, enabled, created_at, updated_at)
SELECT user_id, name, secret, enabled, created_at, updated_at FROM user_mfa WHERE mfa_type = 'totp';
INSERT OR IGNORE INTO user_mfa_passkey(user_id, name, credential_id, credential_json, enabled, direct_login_enabled, created_at, updated_at)
SELECT user_id, name, credential_id, credential_json, enabled, direct_login_enabled, created_at, updated_at FROM user_mfa WHERE mfa_type = 'passkey';
INSERT OR IGNORE INTO user_mfa_email(user_id, mfa_enabled, direct_login_enabled, created_at, updated_at)
SELECT user_id, enabled, direct_login_enabled, created_at, updated_at FROM user_mfa WHERE mfa_type = 'email';

DROP TABLE IF EXISTS user_mfa;
