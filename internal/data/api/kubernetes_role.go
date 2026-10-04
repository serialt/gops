package api

type KubernetesRoleDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesRoleRuleResponse struct {
	APIGroups       []string `json:"api_groups"`
	Resources       []string `json:"resources"`
	ResourceNames   []string `json:"resource_names"`
	Verbs           []string `json:"verbs"`
	NonResourceURLs []string `json:"non_resource_urls"`
}

type KubernetesRoleSupportResponse struct {
	Name            string                               `json:"name"`
	Namespace       string                               `json:"namespace"`
	CreatedAt       string                               `json:"created_at"`
	Age             string                               `json:"age"`
	Owner           string                               `json:"owner"`
	UID             string                               `json:"uid"`
	ResourceVersion string                               `json:"resource_version"`
	Labels          map[string]string                    `json:"labels"`
	Annotations     map[string]string                    `json:"annotations"`
	Rules           []*KubernetesRoleRuleResponse        `json:"rules"`
	Related         []*KubernetesRelatedResourceResponse `json:"related"`
	Events          []*EventInfo                         `json:"events"`
	Description     string                               `json:"description"`
}
