package api

type KubernetesSecretDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesSecretSupportResponse struct {
	Name            string                               `json:"name"`
	Namespace       string                               `json:"namespace"`
	CreatedAt       string                               `json:"created_at"`
	Age             string                               `json:"age"`
	Type            string                               `json:"type"`
	Owner           string                               `json:"owner"`
	UID             string                               `json:"uid"`
	ResourceVersion string                               `json:"resource_version"`
	Labels          map[string]string                    `json:"labels"`
	Annotations     map[string]string                    `json:"annotations"`
	Data            map[string]string                    `json:"data"`
	Size            int64                                `json:"size"`
	Related         []*KubernetesRelatedResourceResponse `json:"related"`
	Events          []*EventInfo                         `json:"events"`
	Description     string                               `json:"description"`
}
