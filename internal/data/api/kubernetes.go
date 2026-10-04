package api

import "time"

type KubernetesClusterListRequest struct {
	PageRequest
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

type KubernetesDeploymentDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesPodLogRequest struct {
	ClusterID  int64  `json:"cluster_id" binding:"required"`
	Namespace  string `json:"namespace" binding:"required"`
	Pod        string `json:"pod" binding:"required"`
	Container  string `json:"container" binding:"required"`
	TailLines  int64  `json:"tail_lines"`
	Timestamps bool   `json:"timestamps"`
}

type KubernetesPodExecRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Pod       string `json:"pod" binding:"required"`
	Container string `json:"container" binding:"required"`
	Command   string `json:"command" binding:"required"`
}

type KubernetesPodLogResponse struct {
	Logs string `json:"logs"`
}
type KubernetesPodExecResponse struct {
	Output string `json:"output"`
	Error  string `json:"error"`
}
type KubernetesPodFileRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Pod       string `json:"pod" binding:"required"`
	Container string `json:"container" binding:"required"`
	Path      string `json:"path" binding:"required"`
}
type KubernetesPodFileResponse struct {
	Name       string `json:"name"`
	Path       string `json:"path"`
	UID        string `json:"uid"`
	GID        string `json:"gid"`
	Size       int64  `json:"size"`
	Mode       string `json:"mode"`
	ModifiedAt int64  `json:"modified_at"`
	Directory  bool   `json:"directory"`
}
type KubernetesPodFileListResponse struct {
	List []*KubernetesPodFileResponse `json:"list"`
}
type KubernetesRelatedResourceResponse struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status"`
}
type KubernetesReplicaSetResponse struct {
	Name          string   `json:"name"`
	Revision      string   `json:"revision"`
	Replicas      int32    `json:"replicas"`
	ReadyReplicas int32    `json:"ready_replicas"`
	Images        []string `json:"images"`
	CreatedAt     string   `json:"created_at"`
	ChangeReason  string   `json:"change_reason"`
}
type KubernetesPodMetricResponse struct {
	Pod    string `json:"pod"`
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}
type KubernetesDeploymentSupportResponse struct {
	Related                 []*KubernetesRelatedResourceResponse `json:"related"`
	History                 []*KubernetesReplicaSetResponse      `json:"history"`
	Metrics                 []*KubernetesPodMetricResponse       `json:"metrics"`
	Owner                   string                               `json:"owner"`
	UID                     string                               `json:"uid"`
	Revision                string                               `json:"revision"`
	MinReadySeconds         int32                                `json:"min_ready_seconds"`
	ProgressDeadlineSeconds *int32                               `json:"progress_deadline_seconds"`
	Description             string                               `json:"description"`
}
type KubernetesClusterTestResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}
type KubernetesClusterSyncResponse struct {
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
	Errors  []string `json:"errors"`
}
type KubernetesClusterConfigResponse struct {
	KubeConfig string `json:"kube_config"`
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
