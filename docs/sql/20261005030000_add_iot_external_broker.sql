-- +goose Up
CREATE TABLE iot_external_broker (
  id INTEGER PRIMARY KEY AUTOINCREMENT, /* 外部 Broker ID */
  name TEXT NOT NULL, /* Broker 名称 */
  url TEXT NOT NULL, /* Broker 连接地址 */
  client_id TEXT NOT NULL, /* MQTT Client ID */
  username TEXT NOT NULL DEFAULT '', /* MQTT 用户名 */
  password TEXT NOT NULL DEFAULT '', /* 加密后的 MQTT 密码 */
  topic_filter TEXT NOT NULL DEFAULT '#', /* 订阅 Topic 过滤器 */
  enabled INTEGER NOT NULL DEFAULT 1, /* 是否启用 */
  insecure_skip_verify INTEGER NOT NULL DEFAULT 0, /* 是否跳过 TLS 证书校验 */
  created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 创建时间 */
  updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP, /* 更新时间 */
  deleted_at DATETIME /* 删除时间 */
);
CREATE UNIQUE INDEX idx_iot_external_broker_name_active ON iot_external_broker(name) WHERE deleted_at IS NULL;
CREATE UNIQUE INDEX idx_iot_external_broker_client_id_active ON iot_external_broker(client_id) WHERE deleted_at IS NULL;

-- +goose Down
DROP TABLE IF EXISTS iot_external_broker;
