-- +goose Up
CREATE UNIQUE INDEX idx_user_mfa_totp_user_id_unique ON user_mfa_totp (user_id);

-- +goose Down
DROP INDEX IF EXISTS idx_user_mfa_totp_user_id_unique;
