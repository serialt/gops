package api

type KubernetesIngressDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesIngressSupportResponse struct {
	Name            string                              `json:"name"`
	Namespace       string                              `json:"namespace"`
	IngressClass    string                              `json:"ingress_class"`
	Hosts           []string                            `json:"hosts"`
	Addresses       []string                            `json:"addresses"`
	Rules           []*KubernetesIngressRuleResponse    `json:"rules"`
	TLS             []*KubernetesIngressTLSResponse     `json:"tls"`
	DefaultBackend  *KubernetesIngressBackendResponse   `json:"default_backend,omitempty"`
	Backends        []*KubernetesIngressServiceResponse `json:"backends"`
	Events          []*EventInfo                        `json:"events"`
	Labels          map[string]string                   `json:"labels"`
	Annotations     map[string]string                   `json:"annotations"`
	Owner           string                              `json:"owner"`
	UID             string                              `json:"uid"`
	ResourceVersion string                              `json:"resource_version"`
	CreatedAt       string                              `json:"created_at"`
	Age             string                              `json:"age"`
	Description     string                              `json:"description"`
}

type KubernetesIngressRuleResponse struct {
	Host  string                           `json:"host"`
	Paths []*KubernetesIngressPathResponse `json:"paths"`
}

type KubernetesIngressPathResponse struct {
	Path        string `json:"path"`
	PathType    string `json:"path_type"`
	ServiceName string `json:"service_name"`
	ServicePort string `json:"service_port"`
}

type KubernetesIngressTLSResponse struct {
	Hosts      []string `json:"hosts"`
	SecretName string   `json:"secret_name"`
}

type KubernetesIngressBackendResponse struct {
	ServiceName string `json:"service_name"`
	ServicePort string `json:"service_port"`
}

type KubernetesIngressServiceResponse struct {
	Name      string `json:"name"`
	Type      string `json:"type"`
	ClusterIP string `json:"cluster_ip"`
	Ports     string `json:"ports"`
}
