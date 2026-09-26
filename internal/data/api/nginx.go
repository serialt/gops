package api

import "time"

type NginxSourceListRequest struct {
	PageRequest
	Keyword string `json:"keyword"`
}
type NginxSourceRequest struct {
	ID              int64  `json:"id"`
	Name            string `json:"name" binding:"required"`
	SourceType      string `json:"source_type"`
	Description     string `json:"description"`
	Status          int64  `json:"status"`
	HostID          int64  `json:"host_id"`
	LogPath         string `json:"log_path"`
	LogFormat       string `json:"log_format"`
	CollectInterval int64  `json:"collect_interval"`
	RetentionDays   int64  `json:"retention_days"`
}
type NginxSourceIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type NginxSourceResponse struct {
	ID              int64     `json:"id"`
	Name            string    `json:"name"`
	SourceType      string    `json:"source_type"`
	Description     string    `json:"description"`
	Status          int64     `json:"status"`
	HostID          int64     `json:"host_id"`
	LogPath         string    `json:"log_path"`
	LogFormat       string    `json:"log_format"`
	CollectInterval int64     `json:"collect_interval"`
	RetentionDays   int64     `json:"retention_days"`
	LastCollectLogs int64     `json:"last_collect_logs"`
	LastError       string    `json:"last_error"`
	UpdatedAt       time.Time `json:"updated_at"`
}
type NginxOverviewResponse struct {
	Sources       int64 `json:"sources"`
	ActiveSources int64 `json:"active_sources"`
	Requests      int64 `json:"requests"`
	ErrorRequests int64 `json:"error_requests"`
}
type NginxAccessLogListRequest struct {
	PageRequest
	SourceID int64  `json:"source_id"`
	Keyword  string `json:"keyword"`
}
type NginxAccessLogResponse struct {
	ID            int64     `json:"id"`
	SourceID      int64     `json:"source_id"`
	OccurredAt    time.Time `json:"occurred_at"`
	RemoteAddr    string    `json:"remote_addr"`
	Method        string    `json:"method"`
	URI           string    `json:"uri"`
	Status        int64     `json:"status"`
	BodyBytesSent int64     `json:"body_bytes_sent"`
	RequestTime   float64   `json:"request_time"`
	UserAgent     string    `json:"user_agent"`
}
