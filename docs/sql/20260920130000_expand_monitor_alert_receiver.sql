-- +goose Up
ALTER TABLE monitor_alert_receiver ADD COLUMN wechat_id VARCHAR(100) NOT NULL DEFAULT ''; -- 企业微信用户 ID
ALTER TABLE monitor_alert_receiver ADD COLUMN dingtalk_id VARCHAR(100) NOT NULL DEFAULT ''; -- 钉钉用户 ID
ALTER TABLE monitor_alert_receiver ADD COLUMN feishu_id VARCHAR(100) NOT NULL DEFAULT ''; -- 飞书用户 ID
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_email BOOLEAN NOT NULL DEFAULT TRUE; -- 是否启用邮件通知
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_webhook BOOLEAN NOT NULL DEFAULT FALSE; -- 是否启用 Webhook 通知
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_wechat BOOLEAN NOT NULL DEFAULT FALSE; -- 是否启用企业微信通知
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_dingtalk BOOLEAN NOT NULL DEFAULT FALSE; -- 是否启用钉钉通知
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_feishu BOOLEAN NOT NULL DEFAULT FALSE; -- 是否启用飞书通知
ALTER TABLE monitor_alert_receiver ADD COLUMN enable_system_msg BOOLEAN NOT NULL DEFAULT TRUE; -- 是否启用系统消息

-- +goose Down
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_system_msg;
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_feishu;
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_dingtalk;
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_wechat;
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_webhook;
ALTER TABLE monitor_alert_receiver DROP COLUMN enable_email;
ALTER TABLE monitor_alert_receiver DROP COLUMN feishu_id;
ALTER TABLE monitor_alert_receiver DROP COLUMN dingtalk_id;
ALTER TABLE monitor_alert_receiver DROP COLUMN wechat_id;
