-- +goose Up
CREATE TABLE user_mfa_totp_new (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL,
  name TEXT NOT NULL DEFAULT '身份验证器',
  secret TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 1,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);

INSERT INTO user_mfa_totp_new (id, user_id, name, secret, enabled, created_at, updated_at)
SELECT id, user_id, '身份验证器', secret, enabled, created_at, updated_at FROM user_mfa_totp;

DROP TABLE user_mfa_totp;
ALTER TABLE user_mfa_totp_new RENAME TO user_mfa_totp;
CREATE INDEX idx_user_mfa_totp_user_id ON user_mfa_totp (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_mfa_totp_user_id;
CREATE TABLE user_mfa_totp_old (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  user_id INTEGER NOT NULL UNIQUE,
  secret TEXT NOT NULL DEFAULT '',
  enabled INTEGER NOT NULL DEFAULT 0,
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES "user" (id) ON DELETE CASCADE
);
INSERT INTO user_mfa_totp_old (user_id, secret, enabled, created_at, updated_at)
SELECT user_id, secret, enabled, created_at, updated_at
FROM user_mfa_totp
WHERE id IN (SELECT MIN(id) FROM user_mfa_totp GROUP BY user_id);
DROP TABLE user_mfa_totp;
ALTER TABLE user_mfa_totp_old RENAME TO user_mfa_totp;
