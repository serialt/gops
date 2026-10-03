package api

import "time"

type TaskJobListRequest struct {
	PageRequest
	Keyword  string `json:"keyword"`
	TaskType string `json:"task_type"`
	Status   string `json:"status"`
}

type TaskJobResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	TemplateID         int64     `json:"template_id"`
	TaskType           string    `json:"task_type"`
	Status             string    `json:"status"`
	TargetHosts        string    `json:"target_hosts"`
	TargetHostsDisplay string    `json:"target_hosts_display"`
	Parameters         string    `json:"parameters"`
	Result             string    `json:"result"`
	ErrorMessage       string    `json:"error_message"`
	CreatedBy          int64     `json:"created_by"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type TaskTemplateListRequest struct {
	PageRequest
	Keyword  string `json:"keyword"`
	Category string `json:"category"`
	Platform string `json:"platform"`
	Status   *int64 `json:"status"`
}

type TaskTemplateRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	Description string `json:"description"`
	Content     string `json:"content" binding:"required"`
	Variables   string `json:"variables"`
	Category    string `json:"category" binding:"required"`
	Platform    string `json:"platform"`
	Timeout     int64  `json:"timeout"`
	SortOrder   int64  `json:"sort_order"`
	Status      *int64 `json:"status" binding:"required"`
}

type TaskTemplateIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type TaskTemplateResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	Description string    `json:"description"`
	Content     string    `json:"content"`
	Variables   string    `json:"variables"`
	Category    string    `json:"category"`
	Platform    string    `json:"platform"`
	Timeout     int64     `json:"timeout"`
	SortOrder   int64     `json:"sort_order"`
	Status      int64     `json:"status"`
	CreatedBy   int64     `json:"created_by"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type TaskExecuteRequest struct {
	Name       string  `json:"name"`
	HostIDs    []int64 `json:"host_ids" binding:"required,min=1"`
	TemplateID int64   `json:"template_id"`
	ScriptType string  `json:"script_type" binding:"required,oneof=shell python"`
	Content    string  `json:"content" binding:"required"`
	Timeout    int64   `json:"timeout"`
}

type TaskHostResultResponse struct {
	HostID   int64  `json:"host_id"`
	HostName string `json:"host_name"`
	HostIP   string `json:"host_ip"`
	Status   string `json:"status"`
	Output   string `json:"output"`
	Error    string `json:"error"`
}

type TaskExecuteResponse struct {
	TaskID  int64                     `json:"task_id"`
	Results []*TaskHostResultResponse `json:"results"`
}
type TaskDistributeResponse struct {
	TaskID  int64                     `json:"task_id"`
	Results []*TaskHostResultResponse `json:"results"`
}
