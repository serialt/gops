package api

type KubernetesServiceAccountDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesServiceAccountSupportResponse struct {
	Name             string                               `json:"name"`
	Namespace        string                               `json:"namespace"`
	CreatedAt        string                               `json:"created_at"`
	Age              string                               `json:"age"`
	Owner            string                               `json:"owner"`
	UID              string                               `json:"uid"`
	ResourceVersion  string                               `json:"resource_version"`
	AutomountToken   string                               `json:"automount_token"`
	Secrets          []string                             `json:"secrets"`
	ImagePullSecrets []string                             `json:"image_pull_secrets"`
	Labels           map[string]string                    `json:"labels"`
	Annotations      map[string]string                    `json:"annotations"`
	Related          []*KubernetesRelatedResourceResponse `json:"related"`
	Events           []*EventInfo                         `json:"events"`
	Description      string                               `json:"description"`
}
