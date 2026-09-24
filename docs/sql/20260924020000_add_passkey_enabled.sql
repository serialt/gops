-- +goose Up
ALTER TABLE user_mfa_passkey ADD COLUMN enabled INTEGER NOT NULL DEFAULT 1;

-- +goose Down
ALTER TABLE user_mfa_passkey DROP COLUMN enabled;
