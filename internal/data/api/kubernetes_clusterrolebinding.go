package api

type KubernetesClusterRoleBindingDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesClusterRoleBindingSupportResponse struct {
	Name            string                                  `json:"name"`
	CreatedAt       string                                  `json:"created_at"`
	Age             string                                  `json:"age"`
	Owner           string                                  `json:"owner"`
	UID             string                                  `json:"uid"`
	ResourceVersion string                                  `json:"resource_version"`
	Labels          map[string]string                       `json:"labels"`
	Annotations     map[string]string                       `json:"annotations"`
	RoleRef         *KubernetesRoleBindingRoleRefResponse   `json:"role_ref"`
	Subjects        []*KubernetesRoleBindingSubjectResponse `json:"subjects"`
	Events          []*EventInfo                            `json:"events"`
	Description     string                                  `json:"description"`
}
