package api

type KubernetesObjectRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Keyword   string `json:"keyword"`
	YAML      string `json:"yaml"`
}

type KubernetesObjectResponse struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Kind      string            `json:"kind"`
	Status    string            `json:"status"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
	CreatedAt string            `json:"created_at"`
}

type KubernetesYAMLResponse struct {
	YAML string `json:"yaml"`
}
