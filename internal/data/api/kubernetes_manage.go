package api

type KubernetesObjectRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Keyword   string `json:"keyword"`
	YAML      string `json:"yaml"`
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
}

type KubernetesResourceValues struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type KubernetesYAMLResponse struct {
	YAML string `json:"yaml"`
}
