package api

type KubernetesStorageDetailRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Resource  string `json:"resource" binding:"required"`
	Namespace string `json:"namespace"`
	Name      string `json:"name" binding:"required"`
}

type KubernetesStorageClaimReference struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
}

type KubernetesStorageSupportResponse struct {
	Kind                 string                               `json:"kind"`
	Name                 string                               `json:"name"`
	Namespace            string                               `json:"namespace"`
	CreatedAt            string                               `json:"created_at"`
	Age                  string                               `json:"age"`
	UID                  string                               `json:"uid"`
	ResourceVersion      string                               `json:"resource_version"`
	Status               string                               `json:"status"`
	StorageClassName     string                               `json:"storage_class_name"`
	VolumeName           string                               `json:"volume_name"`
	Capacity             string                               `json:"capacity"`
	RequestedCapacity    string                               `json:"requested_capacity"`
	AccessModes          []string                             `json:"access_modes"`
	VolumeMode           string                               `json:"volume_mode"`
	ReclaimPolicy        string                               `json:"reclaim_policy"`
	ClaimRef             *KubernetesStorageClaimReference     `json:"claim_ref,omitempty"`
	Provisioner          string                               `json:"provisioner"`
	Parameters           map[string]string                    `json:"parameters"`
	MountOptions         []string                             `json:"mount_options"`
	AllowVolumeExpansion *bool                                `json:"allow_volume_expansion,omitempty"`
	VolumeBindingMode    string                               `json:"volume_binding_mode"`
	Labels               map[string]string                    `json:"labels"`
	Annotations          map[string]string                    `json:"annotations"`
	Related              []*KubernetesRelatedResourceResponse `json:"related"`
	Events               []*EventInfo                         `json:"events"`
	Description          string                               `json:"description"`
}
