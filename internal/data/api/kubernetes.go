package api

import "time"

type KubernetesClusterListRequest struct {
	Keyword string `json:"keyword"`
}
type KubernetesClusterRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Alias       string `json:"alias"`
	APIEndpoint string `json:"api_endpoint"`
	KubeConfig  string `json:"kube_config"`
	Region      string `json:"region"`
	Provider    string `json:"provider"`
	Description string `json:"description"`
	Status      int64  `json:"status"`
}
type KubernetesClusterIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type KubernetesClusterResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Alias       string    `json:"alias"`
	APIEndpoint string    `json:"api_endpoint"`
	Version     string    `json:"version"`
	Status      int64     `json:"status"`
	NodeCount   int64     `json:"node_count"`
	PodCount    int64     `json:"pod_count"`
	Region      string    `json:"region"`
	Provider    string    `json:"provider"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
