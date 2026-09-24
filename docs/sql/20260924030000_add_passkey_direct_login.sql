-- +goose Up
ALTER TABLE user_mfa_passkey ADD COLUMN direct_login_enabled INTEGER NOT NULL DEFAULT 0;

-- +goose Down
ALTER TABLE user_mfa_passkey DROP COLUMN direct_login_enabled;
