package api

type KubernetesRoleBindingDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesRoleBindingSubjectResponse struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

type KubernetesRoleBindingRoleRefResponse struct {
	APIGroup string `json:"api_group"`
	Kind     string `json:"kind"`
	Name     string `json:"name"`
}

type KubernetesRoleBindingSupportResponse struct {
	Name            string                                  `json:"name"`
	Namespace       string                                  `json:"namespace"`
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
