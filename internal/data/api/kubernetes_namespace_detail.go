package api

type KubernetesNamespaceDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesNamespaceResourceResponse struct {
	Kind   string `json:"kind"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

type KubernetesNamespaceSupportResponse struct {
	Name            string                                 `json:"name"`
	Status          string                                 `json:"status"`
	CreatedAt       string                                 `json:"created_at"`
	Age             string                                 `json:"age"`
	ResourceVersion string                                 `json:"resource_version"`
	UID             string                                 `json:"uid"`
	Labels          map[string]string                      `json:"labels"`
	Annotations     map[string]string                      `json:"annotations"`
	Resources       []*KubernetesNamespaceResourceResponse `json:"resources"`
	Pods            []*KubernetesObjectResponse            `json:"pods"`
	Events          []*EventInfo                           `json:"events"`
	Description     string                                 `json:"description"`
}
