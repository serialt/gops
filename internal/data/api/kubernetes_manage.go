package api

type KubernetesObjectRequest struct {
	PageRequest
	ClusterID   int64  `json:"cluster_id" binding:"required"`
	Resource    string `json:"resource" binding:"required"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Keyword     string `json:"keyword"`
	ServiceType string `json:"service_type"`
	YAML        string `json:"yaml"`
	Refresh     bool   `json:"refresh"`
	Wait        bool   `json:"wait"`
	Force       bool   `json:"force"`
}

type KubernetesStatefulSetDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Namespace string `json:"namespace" binding:"required"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesStatefulSetSupportResponse struct {
	Status          string                                     `json:"status"`
	DesiredReplicas int32                                      `json:"desired_replicas"`
	CurrentReplicas int32                                      `json:"current_replicas"`
	ReadyReplicas   int32                                      `json:"ready_replicas"`
	UpdatedReplicas int32                                      `json:"updated_replicas"`
	Pods            []*KubernetesObjectResponse                `json:"pods"`
	Claims          []*KubernetesPersistentVolumeClaimResponse `json:"claims"`
	Events          []*EventInfo                               `json:"events"`
	Description     string                                     `json:"description"`
}

type KubernetesPersistentVolumeClaimResponse struct {
	Name             string   `json:"name"`
	Status           string   `json:"status"`
	VolumeName       string   `json:"volume_name"`
	Capacity         string   `json:"capacity"`
	AccessModes      []string `json:"access_modes"`
	StorageClassName string   `json:"storage_class_name"`
	CreatedAt        string   `json:"created_at"`
}

type KubernetesObjectResponse struct {
	Name             string                    `json:"name"`
	Namespace        string                    `json:"namespace"`
	Kind             string                    `json:"kind"`
	Status           string                    `json:"status"`
	Age              string                    `json:"age"`
	Labels           map[string]string         `json:"labels"`
	CreatedAt        string                    `json:"created_at"`
	Ready            string                    `json:"ready,omitempty"`
	Images           []string                  `json:"images,omitempty"`
	Requests         *KubernetesResourceValues `json:"requests,omitempty"`
	Limits           *KubernetesResourceValues `json:"limits,omitempty"`
	Schedule         string                    `json:"schedule,omitempty"`
	LastScheduleTime string                    `json:"last_schedule_time,omitempty"`
	Suspended        bool                      `json:"suspended,omitempty"`
	Restarts         int64                     `json:"restarts,omitempty"`
	PodIP            string                    `json:"pod_ip,omitempty"`
	Node             string                    `json:"node,omitempty"`
	ServiceType      string                    `json:"service_type,omitempty"`
	ClusterIP        string                    `json:"cluster_ip,omitempty"`
	ExternalIP       string                    `json:"external_ip,omitempty"`
	Ports            []string                  `json:"ports,omitempty"`
	EndpointCount    int64                     `json:"endpoint_count,omitempty"`
	Hosts            []string                  `json:"hosts,omitempty"`
	Addresses        []string                  `json:"addresses,omitempty"`
	PolicyTypes      []string                  `json:"policy_types,omitempty"`
}

type KubernetesResourceValues struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type KubernetesYAMLResponse struct {
	YAML string `json:"yaml"`
}

type KubernetesLabelUpdateRequest struct {
	ClusterID int64             `json:"cluster_id" binding:"required"`
	Resource  string            `json:"resource" binding:"required"`
	Namespace string            `json:"namespace"`
	Name      string            `json:"name" binding:"required"`
	Labels    map[string]string `json:"labels"`
}
