package api

import "time"

type MonitorDomainListRequest struct {
	PageRequest
	Keyword string  `json:"keyword"`
	Status  string  `json:"status"`
	IDs     []int64 `json:"ids"`
}
type MonitorDomainRequest struct {
	ID                int64  `json:"id"`
	Domain            string `json:"domain" binding:"required"`
	Port              int64  `json:"port"`
	RequestClient     string `json:"request_client"`
	CheckInterval     int64  `json:"check_interval"`
	CheckJitter       int64  `json:"check_jitter"`
	EnableSSL         bool   `json:"enable_ssl"`
	VerifyCertificate bool   `json:"verify_certificate"`
	EnableAlert       bool   `json:"enable_alert"`
	ResponseThreshold int64  `json:"response_threshold"`
	SSLExpiryDays     int64  `json:"ssl_expiry_days"`
}
type MonitorDomainIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorCheckHistoryListRequest struct {
	PageRequest
	DomainID int64 `json:"domain_id" binding:"required"`
}
type MonitorCheckHistoryResponse struct {
	ID           int64     `json:"id"`
	DomainID     int64     `json:"domain_id"`
	Status       string    `json:"status"`
	ResponseTime int64     `json:"response_time"`
	SSLValid     bool      `json:"ssl_valid"`
	SSLExpiry    time.Time `json:"ssl_expiry"`
	StatusCode   int64     `json:"status_code"`
	ErrorMessage string    `json:"error_message"`
	CheckedAt    time.Time `json:"checked_at"`
}
type MonitorDomainResponse struct {
	ID                int64     `json:"id"`
	Domain            string    `json:"domain"`
	Port              int64     `json:"port"`
	RequestClient     string    `json:"request_client"`
	Status            string    `json:"status"`
	ResponseTime      int64     `json:"response_time"`
	SSLValid          bool      `json:"ssl_valid"`
	SSLExpiry         time.Time `json:"ssl_expiry"`
	SSLRootCA         string    `json:"ssl_root_ca"`
	CheckInterval     int64     `json:"check_interval"`
	CheckJitter       int64     `json:"check_jitter"`
	EnableSSL         bool      `json:"enable_ssl"`
	VerifyCertificate bool      `json:"verify_certificate"`
	EnableAlert       bool      `json:"enable_alert"`
	ResponseThreshold int64     `json:"response_threshold"`
	SSLExpiryDays     int64     `json:"ssl_expiry_days"`
	LastCheck         time.Time `json:"last_check"`
	LastError         string    `json:"last_error"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type MonitorStatsResponse struct {
	Total    int64 `json:"total"`
	Normal   int64 `json:"normal"`
	Abnormal int64 `json:"abnormal"`
	Paused   int64 `json:"paused"`
}
type MonitorStatsRequest struct {
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
}
type MonitorDomainExcelImportResponse struct {
	Created int `json:"created"`
	Updated int `json:"updated"`
}

type MonitorAlertChannelListRequest struct {
	PageRequest
	Keyword string `json:"keyword"`
}
type MonitorAlertChannelConfig struct {
	SMTPHost        string `json:"smtp_host"`
	SMTPPort        int64  `json:"smtp_port"`
	SMTPUser        string `json:"smtp_user"`
	SMTPPassword    string `json:"smtp_password"`
	FromEmail       string `json:"from_email"`
	FromName        string `json:"from_name"`
	WebhookURL      string `json:"webhook_url"`
	WeChatWebhook   string `json:"wechat_webhook"`
	DingTalkWebhook string `json:"dingtalk_webhook"`
	DingTalkSecret  string `json:"dingtalk_secret"`
	FeishuWebhook   string `json:"feishu_webhook"`
}
type MonitorAlertChannelRequest struct {
	ID          int64                      `json:"id"`
	Name        string                     `json:"name" binding:"required"`
	ChannelType string                     `json:"channel_type" binding:"required"`
	Enabled     bool                       `json:"enabled"`
	Config      *MonitorAlertChannelConfig `json:"config" binding:"required"`
}
type MonitorAlertChannelIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorAlertChannelTestRequest struct {
	ChannelType string                     `json:"channel_type" binding:"required"`
	Config      *MonitorAlertChannelConfig `json:"config" binding:"required"`
	Recipient   string                     `json:"recipient"`
}
type MonitorAlertChannelResponse struct {
	ID          int64                      `json:"id"`
	Name        string                     `json:"name"`
	ChannelType string                     `json:"channel_type"`
	Enabled     bool                       `json:"enabled"`
	Config      *MonitorAlertChannelConfig `json:"config"`
	CreatedAt   time.Time                  `json:"created_at"`
	UpdatedAt   time.Time                  `json:"updated_at"`
}

type MonitorAlertReceiverListRequest struct {
	PageRequest
	Keyword string `json:"keyword"`
}
type MonitorAlertReceiverRequest struct {
	ID              int64  `json:"id"`
	Name            string `json:"name" binding:"required"`
	Email           string `json:"email"`
	Phone           string `json:"phone"`
	Webhook         string `json:"webhook"`
	WechatID        string `json:"wechat_id"`
	DingtalkID      string `json:"dingtalk_id"`
	FeishuID        string `json:"feishu_id"`
	EnableEmail     bool   `json:"enable_email"`
	EnableWebhook   bool   `json:"enable_webhook"`
	EnableWechat    bool   `json:"enable_wechat"`
	EnableDingtalk  bool   `json:"enable_dingtalk"`
	EnableFeishu    bool   `json:"enable_feishu"`
	EnableSystemMsg bool   `json:"enable_system_msg"`
	Enabled         bool   `json:"enabled"`
}
type MonitorAlertReceiverIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorAlertReceiverResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	Webhook         string    `json:"webhook"`
	WechatID        string    `json:"wechat_id"`
	DingtalkID      string    `json:"dingtalk_id"`
	FeishuID        string    `json:"feishu_id"`
	EnableEmail     bool      `json:"enable_email"`
	EnableWebhook   bool      `json:"enable_webhook"`
	EnableWechat    bool      `json:"enable_wechat"`
	EnableDingtalk  bool      `json:"enable_dingtalk"`
	EnableFeishu    bool      `json:"enable_feishu"`
	EnableSystemMsg bool      `json:"enable_system_msg"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

type MonitorAlertLogListRequest struct {
	PageRequest
	DomainID int64  `json:"domain_id"`
	Status   string `json:"status"`
	Keyword  string `json:"keyword"`
	ReceiverName string `json:"receiver_name"`
}
type MonitorAlertLogResponse struct {
	ID            int64     `json:"id"`
	DomainID      int64     `json:"domain_id"`
	AlertType     string    `json:"alert_type"`
	Status        string    `json:"status"`
	Message       string    `json:"message"`
	ChannelType   string    `json:"channel_type"`
	ReceiverNames string    `json:"receiver_names"`
	ErrorMessage  string    `json:"error_message"`
	SentAt        time.Time `json:"sent_at"`
}
