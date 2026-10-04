package api

type KubernetesServiceDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesServiceSupportResponse struct {
	Name            string                               `json:"name"`
	Namespace       string                               `json:"namespace"`
	Type            string                               `json:"type"`
	ClusterIP       string                               `json:"cluster_ip"`
	ExternalName    string                               `json:"external_name"`
	ExternalIPs     []string                             `json:"external_ips"`
	Selector        map[string]string                    `json:"selector"`
	Labels          map[string]string                    `json:"labels"`
	Annotations     map[string]string                    `json:"annotations"`
	Owner           string                               `json:"owner"`
	Ports           []*KubernetesServicePortResponse     `json:"ports"`
	Pods            []*KubernetesObjectResponse          `json:"pods"`
	Endpoints       []*KubernetesServiceEndpointResponse `json:"endpoints"`
	EndpointSlices  []*KubernetesServiceEndpointResponse `json:"endpoint_slices"`
	Events          []*EventInfo                         `json:"events"`
	CreatedAt       string                               `json:"created_at"`
	Age             string                               `json:"age"`
	UID             string                               `json:"uid"`
	ResourceVersion string                               `json:"resource_version"`
	Description     string                               `json:"description"`
}

type KubernetesServicePortResponse struct {
	Name       string `json:"name"`
	Protocol   string `json:"protocol"`
	Port       int32  `json:"port"`
	TargetPort string `json:"target_port"`
	NodePort   int32  `json:"node_port"`
}

type KubernetesServiceEndpointResponse struct {
	Name          string   `json:"name"`
	CreatedAt     string   `json:"created_at"`
	Age           string   `json:"age"`
	AddressType   string   `json:"address_type,omitempty"`
	Addresses     []string `json:"addresses"`
	NotReady      []string `json:"not_ready,omitempty"`
	Ports         []string `json:"ports"`
	EndpointCount int      `json:"endpoint_count"`
}
