package api

type KubernetesResourceListRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Keyword   string `json:"keyword"`
	Namespace string `json:"namespace"`
	Status    string `json:"status"`
	Role      string `json:"role"`
}
type KubernetesClusterStatsRequest struct {
	ClusterID int64 `json:"cluster_id" binding:"required"`
}
type KubernetesNodeUpdateRequest struct {
	ClusterID   int64             `json:"cluster_id" binding:"required"`
	Name        string            `json:"name" binding:"required"`
	Labels      map[string]string `json:"labels"`
	Schedulable bool              `json:"schedulable"`
	Taints      *[]TaintInfo      `json:"taints"`
}
type KubernetesNodeDeleteRequest struct {
	ClusterID int64  `json:"cluster_id" binding:"required"`
	Name      string `json:"name" binding:"required"`
}
type KubernetesNamespaceUpdateRequest struct {
	ClusterID int64             `json:"cluster_id" binding:"required"`
	Name      string            `json:"name" binding:"required"`
	Labels    map[string]string `json:"labels"`
}
type KubernetesNodeResponse struct {
	Name             string            `json:"name"`
	Status           string            `json:"status"`
	Roles            string            `json:"roles"`
	Age              string            `json:"age"`
	Version          string            `json:"version"`
	InternalIP       string            `json:"internalIP"`
	ExternalIP       string            `json:"externalIP,omitempty"`
	OSImage          string            `json:"osImage"`
	KernelVersion    string            `json:"kernelVersion"`
	ContainerRuntime string            `json:"containerRuntime"`
	Labels           map[string]string `json:"labels"`
	Annotations      map[string]string `json:"annotations"`
	// 新增字段
	CPUCapacity    string          `json:"cpuCapacity"`    // CPU容量
	MemoryCapacity string          `json:"memoryCapacity"` // 内存容量
	CPUUsed        int64           `json:"cpuUsed"`        // CPU使用量（毫核）
	MemoryUsed     int64           `json:"memoryUsed"`     // 内存使用量（字节）
	PodCount       int             `json:"podCount"`       // Pod数量
	PodCapacity    int             `json:"podCapacity"`    // Pod容量
	Schedulable    bool            `json:"schedulable"`    // 是否可调度
	TaintCount     int             `json:"taintCount"`     // 污点数量
	Taints         []TaintInfo     `json:"taints"`         // 污点详情
	PodCIDR        string          `json:"podCIDR"`        // Pod CIDR
	ProviderID     string          `json:"providerID"`     // Provider ID
	Conditions     []NodeCondition `json:"conditions"`     // 节点条件
	CreatedAt      string          `json:"created_at"`
}

// TaintInfo 污点信息
type TaintInfo struct {
	Key    string `json:"key"`
	Value  string `json:"value"`
	Effect string `json:"effect"`
}

// NodeCondition 节点条件
type NodeCondition struct {
	Type               string `json:"type"`
	Status             string `json:"status"`
	LastHeartbeatTime  string `json:"lastHeartbeatTime"`
	LastTransitionTime string `json:"lastTransitionTime"`
	Reason             string `json:"reason"`
	Message            string `json:"message"`
}

type KubernetesNamespaceResponse struct {
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
	CreatedAt string            `json:"created_at"`
}
type KubernetesWorkloadResponse struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Ready     string `json:"ready"`
	Status    string `json:"status"`
	CreatedAt string `json:"created_at"`
}
type KubernetesWorkloadStatsResponse struct {
	Deployment  int64 `json:"deployment"`
	StatefulSet int64 `json:"stateful_set"`
	DaemonSet   int64 `json:"daemon_set"`
	Job         int64 `json:"job"`
	CronJob     int64 `json:"cron_job"`
	Pod         int64 `json:"pod"`
}

// PodInfo Pod信息
type PodInfo struct {
	Name       string            `json:"name"`
	Namespace  string            `json:"namespace"`
	Ready      string            `json:"ready"`
	Status     string            `json:"status"`
	Phase      string            `json:"phase"`
	Restarts   int32             `json:"restarts"`
	Age        string            `json:"age"`
	IP         string            `json:"ip"`
	Node       string            `json:"node"`
	Labels     map[string]string `json:"labels"`
	Containers []ContainerInfo   `json:"containers"`
}

// ContainerInfo 容器信息
type ContainerInfo struct {
	Name string `json:"name"`
}

// DeploymentInfo Deployment信息
type DeploymentInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	UpToDate  int32             `json:"upToDate"`
	Available int32             `json:"available"`
	Age       string            `json:"age"`
	Replicas  int32             `json:"replicas"`
	Selector  map[string]string `json:"selector"`
	Labels    map[string]string `json:"labels"`
}

// DaemonSetInfo DaemonSet信息
type DaemonSetInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// StatefulSetInfo StatefulSet信息
type StatefulSetInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// JobInfo Job信息
type JobInfo struct {
	Name      string            `json:"name"`
	Namespace string            `json:"namespace"`
	Ready     string            `json:"ready"`
	Age       string            `json:"age"`
	Labels    map[string]string `json:"labels"`
}

// ClusterStats 集群统计信息
type ClusterStats struct {
	NodeCount         int     `json:"nodeCount"`
	WorkloadCount     int     `json:"workloadCount"` // Deployment + DaemonSet + StatefulSet + Job
	PodCount          int     `json:"podCount"`
	CPUUsage          float64 `json:"cpuUsage"`          // CPU使用率百分比
	MemoryUsage       float64 `json:"memoryUsage"`       // 内存使用率百分比
	CPUCapacity       float64 `json:"cpuCapacity"`       // CPU总核数
	MemoryCapacity    float64 `json:"memoryCapacity"`    // 内存总容量(字节)
	CPUAllocatable    float64 `json:"cpuAllocatable"`    // CPU可分配量
	MemoryAllocatable float64 `json:"memoryAllocatable"` // 内存可分配量(字节)
	CPUUsed           float64 `json:"cpuUsed"`           // CPU已使用量
	MemoryUsed        float64 `json:"memoryUsed"`        // 内存已使用量(字节)
}

// ClusterNetworkInfo 集群网络信息
type ClusterNetworkInfo struct {
	ServiceCIDR      string `json:"serviceCIDR"`      // Service CIDR
	PodCIDR          string `json:"podCIDR"`          // Pod CIDR
	APIServerAddress string `json:"apiServerAddress"` // API Server 地址
	NetworkPlugin    string `json:"networkPlugin"`    // 网络插件
	ProxyMode        string `json:"proxyMode"`        // 服务转发模式
	DNSService       string `json:"dnsService"`       // DNS 服务
}

// ClusterComponentInfo 集群组件信息
type ClusterComponentInfo struct {
	Components []ComponentInfo `json:"components"` // 控制平面组件
	Runtime    RuntimeInfo     `json:"runtime"`    // 运行时信息
	Storage    []StorageInfo   `json:"storage"`    // 存储信息
}

// ComponentInfo 组件信息
type ComponentInfo struct {
	Name    string `json:"name"`    // 组件名称
	Version string `json:"version"` // 版本
	Status  string `json:"status"`  // 状态
}

// RuntimeInfo 运行时信息
type RuntimeInfo struct {
	ContainerRuntime string `json:"containerRuntime"` // 容器运行时
	Version          string `json:"version"`          // 版本
}

// StorageInfo 存储信息
type StorageInfo struct {
	Name          string `json:"name"`          // 存储名称
	Provisioner   string `json:"provisioner"`   // Provisioner
	ReclaimPolicy string `json:"reclaimPolicy"` // 回收策略
}

// NetworkPolicyInfo 网络策略信息
type NetworkPolicyInfo struct {
	Name      string            `json:"name"`      // 名称
	Namespace string            `json:"namespace"` // 命名空间
	NetworkID string            `json:"networkID"` // 网络ID (UID)
	Type      string            `json:"type"`      // 类型
	Status    string            `json:"status"`    // 状态
	Age       string            `json:"age"`       // 创建时间
	Labels    map[string]string `json:"labels"`    // 标签
}

// EventInfo 事件信息
type EventInfo struct {
	Type           string             `json:"type"`           // 事件类型: Normal, Warning
	Reason         string             `json:"reason"`         // 原因
	Message        string             `json:"message"`        // 消息
	Source         string             `json:"source"`         // 来源
	Count          int32              `json:"count"`          // 次数
	FirstTimestamp string             `json:"firstTimestamp"` // 首次发生时间
	LastTimestamp  string             `json:"lastTimestamp"`  // 最后发生时间
	InvolvedObject InvolvedObjectInfo `json:"involvedObject"` // 关联对象
}

// InvolvedObjectInfo 关联对象信息
type InvolvedObjectInfo struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace,omitempty"`
}
