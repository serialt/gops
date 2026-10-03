-- +goose Up
-- 为主机增加可选的 VNC 登录凭据。

ALTER TABLE host ADD COLUMN vnc_credential_id BIGINT REFERENCES credential(id) ON DELETE SET NULL /* VNC凭据ID */;
ALTER TABLE host ADD COLUMN vnc_port INTEGER /* VNC端口 */ NOT NULL DEFAULT 5900;

-- +goose Down
ALTER TABLE host DROP COLUMN vnc_port;
ALTER TABLE host DROP COLUMN vnc_credential_id;
