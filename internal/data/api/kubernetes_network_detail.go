package api

type KubernetesNetworkDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesNetworkPolicySupportResponse struct {
	Name            string                                 `json:"name"`
	Namespace       string                                 `json:"namespace"`
	PodSelector     map[string]string                      `json:"pod_selector"`
	PodSelectorText string                                 `json:"pod_selector_text"`
	PolicyTypes     []string                               `json:"policy_types"`
	Ingress         []*KubernetesNetworkPolicyRuleResponse `json:"ingress"`
	Egress          []*KubernetesNetworkPolicyRuleResponse `json:"egress"`
	Pods            []*KubernetesObjectResponse            `json:"pods"`
	Events          []*EventInfo                           `json:"events"`
	Labels          map[string]string                      `json:"labels"`
	Annotations     map[string]string                      `json:"annotations"`
	Owner           string                                 `json:"owner"`
	UID             string                                 `json:"uid"`
	CreatedAt       string                                 `json:"created_at"`
	Age             string                                 `json:"age"`
	Description     string                                 `json:"description"`
}

type KubernetesNetworkPolicyRuleResponse struct {
	Peers []*KubernetesNetworkPolicyPeerResponse `json:"peers"`
	Ports []*KubernetesNetworkPolicyPortResponse `json:"ports"`
}

type KubernetesNetworkPolicyPeerResponse struct {
	PodSelector           map[string]string `json:"pod_selector,omitempty"`
	PodSelectorText       string            `json:"pod_selector_text,omitempty"`
	NamespaceSelector     map[string]string `json:"namespace_selector,omitempty"`
	NamespaceSelectorText string            `json:"namespace_selector_text,omitempty"`
	IPBlockCIDR           string            `json:"ip_block_cidr,omitempty"`
	IPBlockExcept         []string          `json:"ip_block_except,omitempty"`
}

type KubernetesNetworkPolicyPortResponse struct {
	Protocol string `json:"protocol"`
	Port     string `json:"port"`
	EndPort  *int32 `json:"end_port,omitempty"`
}

type KubernetesEndpointsSupportResponse struct {
	Name        string                               `json:"name"`
	Namespace   string                               `json:"namespace"`
	Addresses   []*KubernetesEndpointAddressResponse `json:"addresses"`
	NotReady    []*KubernetesEndpointAddressResponse `json:"not_ready"`
	Ports       []*KubernetesServicePortResponse     `json:"ports"`
	Pods        []*KubernetesObjectResponse          `json:"pods"`
	Events      []*EventInfo                         `json:"events"`
	Labels      map[string]string                    `json:"labels"`
	Annotations map[string]string                    `json:"annotations"`
	Owner       string                               `json:"owner"`
	UID         string                               `json:"uid"`
	CreatedAt   string                               `json:"created_at"`
	Age         string                               `json:"age"`
	Description string                               `json:"description"`
}

type KubernetesEndpointAddressResponse struct {
	IP         string `json:"ip"`
	Hostname   string `json:"hostname"`
	NodeName   string `json:"node_name"`
	TargetKind string `json:"target_kind"`
	TargetName string `json:"target_name"`
}
