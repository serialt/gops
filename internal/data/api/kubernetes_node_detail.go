package api

type KubernetesNodeDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesNodeSupportResponse struct {
	Node        *KubernetesNodeResponse     `json:"node"`
	Pods        []*KubernetesObjectResponse `json:"pods"`
	Events      []*EventInfo                `json:"events"`
	Description string                      `json:"description"`
}
