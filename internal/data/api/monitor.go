package api

import "time"

type MonitorDomainListRequest struct {
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
}
type MonitorDomainRequest struct {
	ID                int64  `json:"id"`
	Domain            string `json:"domain" binding:"required"`
	Port              int64  `json:"port"`
	CheckInterval     int64  `json:"check_interval"`
	EnableSSL         bool   `json:"enable_ssl"`
	EnableAlert       bool   `json:"enable_alert"`
	ResponseThreshold int64  `json:"response_threshold"`
	SSLExpiryDays     int64  `json:"ssl_expiry_days"`
}
type MonitorDomainIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorDomainResponse struct {
	ID                int64     `json:"id"`
	Domain            string    `json:"domain"`
	Port              int64     `json:"port"`
	Status            string    `json:"status"`
	ResponseTime      int64     `json:"response_time"`
	SSLValid          bool      `json:"ssl_valid"`
	SSLExpiry         time.Time `json:"ssl_expiry"`
	CheckInterval     int64     `json:"check_interval"`
	EnableSSL         bool      `json:"enable_ssl"`
	EnableAlert       bool      `json:"enable_alert"`
	ResponseThreshold int64     `json:"response_threshold"`
	SSLExpiryDays     int64     `json:"ssl_expiry_days"`
	LastCheck         time.Time `json:"last_check"`
	LastError         string    `json:"last_error"`
	UpdatedAt         time.Time `json:"updated_at"`
}
type MonitorStatsResponse struct {
	Total    int64 `json:"total"`
	Normal   int64 `json:"normal"`
	Abnormal int64 `json:"abnormal"`
	Paused   int64 `json:"paused"`
}

type MonitorAlertChannelListRequest struct {
	Keyword string `json:"keyword"`
}
type MonitorAlertChannelRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	ChannelType string `json:"channel_type" binding:"required"`
	Enabled     bool   `json:"enabled"`
	Config      string `json:"config"`
}
type MonitorAlertChannelIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorAlertChannelResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	ChannelType string    `json:"channel_type"`
	Enabled     bool      `json:"enabled"`
	Config      string    `json:"config"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MonitorAlertReceiverListRequest struct {
	Keyword string `json:"keyword"`
}
type MonitorAlertReceiverRequest struct {
	ID      int64  `json:"id"`
	Name    string `json:"name" binding:"required"`
	Email   string `json:"email"`
	Phone   string `json:"phone"`
	Webhook string `json:"webhook"`
	Enabled bool   `json:"enabled"`
}
type MonitorAlertReceiverIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type MonitorAlertReceiverResponse struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	Webhook   string    `json:"webhook"`
	Enabled   bool      `json:"enabled"`
	UpdatedAt time.Time `json:"updated_at"`
}

type MonitorAlertLogListRequest struct {
	DomainID int64  `json:"domain_id"`
	Status   string `json:"status"`
}
type MonitorAlertLogResponse struct {
	ID           int64     `json:"id"`
	DomainID     int64     `json:"domain_id"`
	AlertType    string    `json:"alert_type"`
	Status       string    `json:"status"`
	Message      string    `json:"message"`
	ChannelType  string    `json:"channel_type"`
	ErrorMessage string    `json:"error_message"`
	SentAt       time.Time `json:"sent_at"`
}
