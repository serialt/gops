package handler

import (
	"context"
	"errors"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
	"mime"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type KubernetesHandler struct{ biz *biz.KubernetesBiz }

func NewKubernetesHandler(b *biz.KubernetesBiz) *KubernetesHandler { return &KubernetesHandler{b} }

func (h *KubernetesHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/kubernetes/clusters/list", h.ListClusters)
	g.POST("/kubernetes/clusters/save", h.SaveCluster)
	g.POST("/kubernetes/clusters/delete", h.DeleteCluster)
	g.POST("/kubernetes/clusters/test", h.TestCluster)
	g.POST("/kubernetes/clusters/sync", h.SyncClusters)
	g.POST("/kubernetes/clusters/config", h.GetClusterConfig)
	g.POST("/kubernetes/nodes/list", h.ListNodes)
	g.POST("/kubernetes/nodes/support", h.GetNodeSupport)
	g.POST("/kubernetes/nodes/update", h.UpdateNode)
	g.POST("/kubernetes/nodes/delete", h.DeleteNode)
	g.POST("/kubernetes/nodes/drain", h.DrainNode)
	g.POST("/kubernetes/namespaces/list", h.ListNamespaces)
	g.POST("/kubernetes/namespaces/support", h.GetNamespaceSupport)
	g.POST("/kubernetes/namespaces/update", h.UpdateNamespace)
	g.POST("/kubernetes/workloads/list", h.ListWorkloads)
	g.POST("/kubernetes/workloads/stats", h.WorkloadStats)
	g.POST("/kubernetes/pods/list", h.ListPods)
	g.POST("/kubernetes/stats/get", h.GetClusterStats)
	g.POST("/kubernetes/events/list", h.ListEvents)
	g.POST("/kubernetes/ingresses/support", h.GetIngressSupport)
	g.POST("/kubernetes/networkpolicies/support", h.GetNetworkPolicySupport)
	g.POST("/kubernetes/endpoints/support", h.GetEndpointsSupport)
	g.POST("/kubernetes/services/support", h.GetServiceSupport)
	g.POST("/kubernetes/deployments/support", h.GetDeploymentSupport)
	g.POST("/kubernetes/statefulsets/support", h.GetStatefulSetSupport)
	g.POST("/kubernetes/daemonsets/support", h.GetDaemonSetSupport)
	g.POST("/kubernetes/jobs/support", h.GetJobSupport)
	g.POST("/kubernetes/cronjobs/support", h.GetCronJobSupport)
	g.POST("/kubernetes/pods/logs", h.GetPodLogs)
	g.POST("/kubernetes/pods/exec", h.ExecPodCommand)
	g.POST("/kubernetes/pods/files/list", h.ListPodFiles)
	g.POST("/kubernetes/pods/files/upload", h.UploadPodFile)
	g.POST("/kubernetes/pods/files/download", h.DownloadPodFile)
	g.POST("/kubernetes/resources/list", h.ListKubernetesObjects)
	g.POST("/kubernetes/resources/get", h.GetKubernetesObject)
	g.POST("/kubernetes/resources/template", h.GetKubernetesResourceTemplate)
	g.POST("/kubernetes/resources/apply", h.ApplyKubernetesObject)
	g.POST("/kubernetes/resources/delete", h.DeleteKubernetesObject)
	g.POST("/kubernetes/resources/labels/update", h.UpdateKubernetesObjectLabels)
	g.POST("/kubernetes/inspections/start", h.StartInspection)
	g.POST("/kubernetes/inspections/progress", h.GetInspectionProgress)
	g.POST("/kubernetes/inspections/result", h.GetInspectionResult)
	g.POST("/kubernetes/inspections/history", h.ListInspectionHistory)
	g.POST("/kubernetes/inspections/delete", h.DeleteInspection)

}

// GetNetworkPolicySupport POST /api/v1/kubernetes/networkpolicies/support 查询 NetworkPolicy 规则、目标 Pod 和事件。
func (h *KubernetesHandler) GetNetworkPolicySupport(c *gin.Context) {
	req := new(api.KubernetesNetworkDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid network policy query")
		return
	}
	data, err := h.biz.NetworkPolicySupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetEndpointsSupport POST /api/v1/kubernetes/endpoints/support 查询 Endpoints 地址、端口、目标 Pod 和事件。
func (h *KubernetesHandler) GetEndpointsSupport(c *gin.Context) {
	req := new(api.KubernetesNetworkDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid endpoints query")
		return
	}
	data, err := h.biz.EndpointsSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetIngressSupport POST /api/v1/kubernetes/ingresses/support 查询 Ingress 规则、TLS、后端 Service 和事件。
func (h *KubernetesHandler) GetIngressSupport(c *gin.Context) {
	req := new(api.KubernetesIngressDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid ingress query")
		return
	}
	data, err := h.biz.IngressSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetNodeSupport POST /api/v1/kubernetes/nodes/support 查询 Node 详情、所在 Pods 和事件。
func (h *KubernetesHandler) GetNodeSupport(c *gin.Context) {
	req := new(api.KubernetesNodeDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid node query")
		return
	}
	data, err := h.biz.NodeSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetNamespaceSupport POST /api/v1/kubernetes/namespaces/support 查询 Namespace 资源、Pod 和事件详情。
func (h *KubernetesHandler) GetNamespaceSupport(c *gin.Context) {
	req := new(api.KubernetesNamespaceDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid namespace query")
		return
	}
	data, err := h.biz.NamespaceSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetServiceSupport POST /api/v1/kubernetes/services/support 查询 Service 关联的 Pod、Endpoints、EndpointSlices 和事件。
func (h *KubernetesHandler) GetServiceSupport(c *gin.Context) {
	req := new(api.KubernetesServiceDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid service query")
		return
	}
	data, err := h.biz.ServiceSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetJobSupport POST /api/v1/kubernetes/jobs/support 查询 Job 关联的 Pod 和事件。
func (h *KubernetesHandler) GetJobSupport(c *gin.Context) {
	req := new(api.KubernetesJobDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid job query")
		return
	}
	data, err := h.biz.JobSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetCronJobSupport POST /api/v1/kubernetes/cronjobs/support 查询 CronJob 关联的 Job、Pod 和事件。
func (h *KubernetesHandler) GetCronJobSupport(c *gin.Context) {
	req := new(api.KubernetesCronJobDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid cronjob query")
		return
	}
	data, err := h.biz.CronJobSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetDaemonSetSupport POST /api/v1/kubernetes/daemonsets/support 查询 DaemonSet 关联的 Pod 和事件。
func (h *KubernetesHandler) GetDaemonSetSupport(c *gin.Context) {
	req := new(api.KubernetesDaemonSetDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid daemonset query")
		return
	}
	data, err := h.biz.DaemonSetSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetStatefulSetSupport POST /api/v1/kubernetes/statefulsets/support 查询 StatefulSet 关联的 Pod、PVC 和事件。
func (h *KubernetesHandler) GetStatefulSetSupport(c *gin.Context) {
	req := new(api.KubernetesStatefulSetDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid statefulset query")
		return
	}
	data, err := h.biz.StatefulSetSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetDeploymentSupport POST /api/v1/kubernetes/deployments/support 查询 Deployment 关联资源、历史和监控指标。
func (h *KubernetesHandler) GetDeploymentSupport(c *gin.Context) {
	req := new(api.KubernetesDeploymentDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid deployment query")
		return
	}
	data, err := h.biz.DeploymentSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetPodLogs POST /api/v1/kubernetes/pods/logs 查询 Pod 容器日志。
func (h *KubernetesHandler) GetPodLogs(c *gin.Context) {
	req := new(api.KubernetesPodLogRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid pod log query")
		return
	}
	data, err := h.biz.PodLogs(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ExecPodCommand POST /api/v1/kubernetes/pods/exec 在 Pod 容器中执行命令。
func (h *KubernetesHandler) ExecPodCommand(c *gin.Context) {
	req := new(api.KubernetesPodExecRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid pod exec request")
		return
	}
	data, err := h.biz.PodExec(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ListPodFiles POST /api/v1/kubernetes/pods/files/list 查询 Pod 容器目录。
func (h *KubernetesHandler) ListPodFiles(c *gin.Context) {
	req := new(api.KubernetesPodFileRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid pod file query")
		return
	}
	data, err := h.biz.PodFiles(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// UploadPodFile POST /api/v1/kubernetes/pods/files/upload 上传文件到 Pod 容器。
func (h *KubernetesHandler) UploadPodFile(c *gin.Context) {
	clusterID, err := strconv.ParseInt(c.PostForm("cluster_id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}
	req := &api.KubernetesPodFileRequest{ClusterID: clusterID, Namespace: c.PostForm("namespace"), Pod: c.PostForm("pod"), Container: c.PostForm("container"), Path: c.PostForm("path")}
	header, err := c.FormFile("file")
	if err != nil {
		response.Error(c, 400, "file is required")
		return
	}
	file, err := header.Open()
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	defer file.Close()
	if err = h.biz.UploadPodFile(req, header.Filename, file); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// DownloadPodFile POST /api/v1/kubernetes/pods/files/download 下载 Pod 容器文件。
func (h *KubernetesHandler) DownloadPodFile(c *gin.Context) {
	req := new(api.KubernetesPodFileRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid pod file query")
		return
	}
	data, err := h.biz.DownloadPodFile(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	filename := filepath.Base(req.Path)
	c.Header("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": filename}))
	c.Data(200, "application/octet-stream", data)
}

func (h *KubernetesHandler) StartStatusScheduler(ctx context.Context, interval time.Duration) {
	h.biz.StartStatusScheduler(ctx, interval)
}

// ListClusters POST /api/v1/kubernetes/clusters/list 查询 Kubernetes 集群。
func (h *KubernetesHandler) ListClusters(c *gin.Context) {
	r := new(api.KubernetesClusterListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster query")
		return
	}
	d, e := h.biz.List(r)
	if e != nil {
		response.Error(c, 500, e.Error())
		return
	}
	response.OK(c, d)
}

// SaveCluster POST /api/v1/kubernetes/clusters/save 保存 Kubernetes 集群。
func (h *KubernetesHandler) SaveCluster(c *gin.Context) {
	r := new(api.KubernetesClusterRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster data")
		return
	}
	d, e := h.biz.Save(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// DeleteCluster POST /api/v1/kubernetes/clusters/delete 删除 Kubernetes 集群。
func (h *KubernetesHandler) DeleteCluster(c *gin.Context) {
	r := new(api.KubernetesClusterIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}
	if e := h.biz.Delete(r); e != nil {
		if errors.Is(e, biz.ErrNotFound) {
			response.Error(c, 404, "cluster not found")
		} else {
			response.Error(c, 400, e.Error())
		}
		return
	}
	response.OK(c, nil)
}

// TestCluster POST /api/v1/kubernetes/clusters/test 测试 Kubernetes 集群连接并同步状态。
func (h *KubernetesHandler) TestCluster(c *gin.Context) {
	r := new(api.KubernetesClusterIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}
	d, e := h.biz.Test(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// SyncClusters POST /api/v1/kubernetes/clusters/sync 同步全部 Kubernetes 集群状态。
func (h *KubernetesHandler) SyncClusters(c *gin.Context) {
	d, e := h.biz.SyncAll(c.Request.Context())
	if e != nil {
		response.Error(c, 500, e.Error())
		return
	}
	response.OK(c, d)
}

// GetClusterConfig POST /api/v1/kubernetes/clusters/config 获取 Kubernetes 集群凭证。
func (h *KubernetesHandler) GetClusterConfig(c *gin.Context) {
	r := new(api.KubernetesClusterIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}
	d, e := h.biz.Config(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// ListNodes POST /api/v1/kubernetes/nodes/list 查询 Kubernetes 节点。
func (h *KubernetesHandler) ListNodes(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid node query")
		return
	}
	d, e := h.biz.Nodes(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// ListNamespaces POST /api/v1/kubernetes/namespaces/list 查询 Kubernetes 命名空间。
func (h *KubernetesHandler) ListNamespaces(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid namespace query")
		return
	}
	d, e := h.biz.Namespaces(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// ListWorkloads POST /api/v1/kubernetes/workloads/list 查询 Kubernetes 工作负载。
func (h *KubernetesHandler) ListWorkloads(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid workload query")
		return
	}
	d, e := h.biz.Workloads(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// WorkloadStats POST /api/v1/kubernetes/workloads/stats 查询 Kubernetes 工作负载统计。
func (h *KubernetesHandler) WorkloadStats(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid workload stats query")
		return
	}
	d, e := h.biz.WorkloadStats(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// ListPods POST /api/v1/kubernetes/pods/list 查询 Kubernetes Pod。
func (h *KubernetesHandler) ListPods(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid pod query")
		return
	}
	d, e := h.biz.Pods(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// UpdateNode POST /api/v1/kubernetes/nodes/update 更新 Kubernetes 节点配置。
func (h *KubernetesHandler) UpdateNode(c *gin.Context) {
	r := new(api.KubernetesNodeUpdateRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid node update request")
		return
	}
	d, e := h.biz.UpdateNode(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// DeleteNode POST /api/v1/kubernetes/nodes/delete 删除 Kubernetes 节点。
func (h *KubernetesHandler) DeleteNode(c *gin.Context) {
	r := new(api.KubernetesNodeDeleteRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid node delete request")
		return
	}
	if e := h.biz.DeleteNode(r); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, nil)
}

// DrainNode POST /api/v1/kubernetes/nodes/drain 排空 Kubernetes 节点。
func (h *KubernetesHandler) DrainNode(c *gin.Context) {
	r := new(api.KubernetesNodeDeleteRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid node drain request")
		return
	}
	if e := h.biz.DrainNode(r); e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, nil)
}

// UpdateNamespace POST /api/v1/kubernetes/namespaces/update 更新 Kubernetes 命名空间标签。
func (h *KubernetesHandler) UpdateNamespace(c *gin.Context) {
	r := new(api.KubernetesNamespaceUpdateRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid namespace update request")
		return
	}
	d, e := h.biz.UpdateNamespace(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// GetClusterStats POST /api/v1/kubernetes/stats/get 查询 Kubernetes 集群统计。
func (h *KubernetesHandler) GetClusterStats(c *gin.Context) {
	r := new(api.KubernetesClusterStatsRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid cluster id")
		return
	}
	d, e := h.biz.Stats(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}

// ListEvents POST /api/v1/kubernetes/events/list 查询 Kubernetes 事件。
func (h *KubernetesHandler) ListEvents(c *gin.Context) {
	r := new(api.KubernetesResourceListRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid event query")
		return
	}
	d, e := h.biz.Events(r)
	if e != nil {
		response.Error(c, 400, e.Error())
		return
	}
	response.OK(c, d)
}
