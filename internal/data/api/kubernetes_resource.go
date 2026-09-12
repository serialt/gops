package api

type KubernetesResourceListRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Keyword   string `json:"keyword"`
	Namespace string `json:"namespace"`
}
type KubernetesNodeResponse struct {
	Name       string `json:"name"`
	Status     string `json:"status"`
	Roles      string `json:"roles"`
	Version    string `json:"version"`
	InternalIP string `json:"internal_ip"`
	CreatedAt  string `json:"created_at"`
}
type KubernetesNamespaceResponse struct {
	Name      string `json:"name"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
type KubernetesWorkloadResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
