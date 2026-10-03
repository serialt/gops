-- +goose Up
-- 主机的 SSH/VNC 端口为可选配置，NULL 表示未配置对应连接。

ALTER TABLE host ADD COLUMN ssh_port_nullable INTEGER /* SSH端口 */;
UPDATE host SET ssh_port_nullable = CASE WHEN credential_id IS NULL THEN NULL ELSE NULLIF(port, 0) END;
ALTER TABLE host DROP COLUMN port;
ALTER TABLE host RENAME COLUMN ssh_port_nullable TO port;

ALTER TABLE host ADD COLUMN vnc_port_nullable INTEGER /* VNC端口 */;
UPDATE host SET vnc_port_nullable = CASE WHEN vnc_credential_id IS NULL THEN NULL ELSE NULLIF(vnc_port, 0) END;
ALTER TABLE host DROP COLUMN vnc_port;
ALTER TABLE host RENAME COLUMN vnc_port_nullable TO vnc_port;

-- +goose Down
ALTER TABLE host ADD COLUMN ssh_port_required INTEGER /* SSH端口 */ NOT NULL DEFAULT 22;
UPDATE host SET ssh_port_required = COALESCE(port, 22);
ALTER TABLE host DROP COLUMN port;
ALTER TABLE host RENAME COLUMN ssh_port_required TO port;

ALTER TABLE host ADD COLUMN vnc_port_required INTEGER /* VNC端口 */ NOT NULL DEFAULT 5900;
UPDATE host SET vnc_port_required = COALESCE(vnc_port, 5900);
ALTER TABLE host DROP COLUMN vnc_port;
ALTER TABLE host RENAME COLUMN vnc_port_required TO vnc_port;
