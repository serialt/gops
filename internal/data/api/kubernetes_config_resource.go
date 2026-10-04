package api

type KubernetesConfigResourceSupportRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesConfigResourcePropertyResponse struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type KubernetesConfigResourceSupportResponse struct {
	Kind            string                                      `json:"kind"`
	Name            string                                      `json:"name"`
	Namespace       string                                      `json:"namespace"`
	CreatedAt       string                                      `json:"created_at"`
	Age             string                                      `json:"age"`
	UID             string                                      `json:"uid"`
	ResourceVersion string                                      `json:"resource_version"`
	Status          string                                      `json:"status"`
	Owner           string                                      `json:"owner"`
	Labels          map[string]string                           `json:"labels"`
	Annotations     map[string]string                           `json:"annotations"`
	Properties      []*KubernetesConfigResourcePropertyResponse `json:"properties"`
	Related         []*KubernetesRelatedResourceResponse        `json:"related"`
	Events          []*EventInfo                                `json:"events"`
	Description     string                                      `json:"description"`
}
