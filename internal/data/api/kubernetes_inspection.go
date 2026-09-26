package api

import "time"

type KubernetesInspectionStartRequest struct {
	ClusterID int64 `json:"cluster_id" binding:"required"`
	UserID    int64 `json:"-"`
}

type KubernetesInspectionIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type KubernetesInspectionHistoryRequest struct {
	PageRequest
	ClusterID int64 `json:"cluster_id"`
}

type KubernetesInspectionStartResponse struct {
	ID int64 `json:"id"`
}

type KubernetesInspectionProgressResponse struct {
	ID          int64  `json:"id"`
	Status      string `json:"status"`
	Progress    int    `json:"progress"`
	CurrentStep string `json:"current_step"`
}

type KubernetesInspectionCheck struct {
	Category   string `json:"category"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	Value      string `json:"value"`
	Expected   string `json:"expected"`
	Detail     string `json:"detail"`
	Suggestion string `json:"suggestion"`
}

type KubernetesInspectionSection struct {
	Key     string                        `json:"key"`
	Title   string                        `json:"title"`
	Metrics []*KubernetesInspectionMetric `json:"metrics,omitempty"`
	Details []map[string]any              `json:"details,omitempty"`
	Checks  []*KubernetesInspectionCheck  `json:"checks"`
}

type KubernetesInspectionMetric struct {
	Label string `json:"label"`
	Value string `json:"value"`
	Tone  string `json:"tone,omitempty"`
}

type KubernetesInspectionSummary struct {
	Total    int `json:"total"`
	Passed   int `json:"passed"`
	Warning  int `json:"warning"`
	Failed   int `json:"failed"`
	Duration int `json:"duration"`
}

type KubernetesInspectionReport struct {
	ClusterID   int64                          `json:"cluster_id"`
	ClusterName string                         `json:"cluster_name"`
	Score       int                            `json:"score"`
	Summary     *KubernetesInspectionSummary   `json:"summary"`
	Sections    []*KubernetesInspectionSection `json:"sections"`
}

type KubernetesInspectionHistoryResponse struct {
	ID           int64     `json:"id"`
	ClusterID    int64     `json:"cluster_id"`
	ClusterName  string    `json:"cluster_name"`
	Status       string    `json:"status"`
	Score        int64     `json:"score"`
	CheckCount   int64     `json:"check_count"`
	PassCount    int64     `json:"pass_count"`
	WarningCount int64     `json:"warning_count"`
	FailCount    int64     `json:"fail_count"`
	Duration     int64     `json:"duration"`
	StartedAt    time.Time `json:"started_at"`
	EndedAt      time.Time `json:"ended_at"`
}

type KubernetesInspectionResultResponse struct {
	Inspection *KubernetesInspectionHistoryResponse `json:"inspection"`
	Report     *KubernetesInspectionReport          `json:"report"`
}
