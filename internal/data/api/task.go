package api

import "time"

type TaskListRequest struct {
	PageRequest
	Keyword string `json:"keyword"`
}
type TaskJobRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	TemplateID  int64  `json:"template_id"`
	TaskType    string `json:"task_type"`
	TargetHosts string `json:"target_hosts"`
	Parameters  string `json:"parameters"`
}
type TaskJobResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	TemplateID  int64     `json:"template_id"`
	TaskType    string    `json:"task_type"`
	Status      string    `json:"status"`
	TargetHosts string    `json:"target_hosts"`
	Parameters  string    `json:"parameters"`
	CreatedAt   time.Time `json:"created_at"`
}
