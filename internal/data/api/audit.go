package api

import "time"

type AuditListRequest struct {
	PageRequest
	Keyword   string `json:"keyword"`
	Username  string `json:"username"`
	Module    string `json:"module"`
	Action    string `json:"action"`
	Status    string `json:"status"`
	LoginType string `json:"login_type"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}
type AuditDeleteRequest struct {
	IDs []int64 `json:"ids" binding:"required,min=1,dive,gt=0"`
}
type AuditOperationLogResponse struct {
	ID          int64     `json:"id"`
	Username    string    `json:"username"`
	Module      string    `json:"module"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Status      int64     `json:"status"`
	CostTime    int64     `json:"cost_time"`
	IP          string    `json:"ip"`
	CreatedAt   time.Time `json:"created_at"`
}
type AuditLoginLogResponse struct {
	ID        int64     `json:"id"`
	Username  string    `json:"username"`
	LoginType string    `json:"login_type"`
	Status    string    `json:"status"`
	IP        string    `json:"ip"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}
