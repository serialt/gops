package api

type KubernetesClusterRoleDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesClusterRoleSupportResponse struct {
	Name            string                               `json:"name"`
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
