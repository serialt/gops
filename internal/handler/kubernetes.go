package handler

import (
	"context"
	"errors"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
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
	g.POST("/kubernetes/nodes/update", h.UpdateNode)
	g.POST("/kubernetes/nodes/delete", h.DeleteNode)
	g.POST("/kubernetes/nodes/drain", h.DrainNode)
	g.POST("/kubernetes/namespaces/list", h.ListNamespaces)
	g.POST("/kubernetes/namespaces/update", h.UpdateNamespace)
	g.POST("/kubernetes/workloads/list", h.ListWorkloads)
	g.POST("/kubernetes/workloads/stats", h.WorkloadStats)
	g.POST("/kubernetes/pods/list", h.ListPods)
	g.POST("/kubernetes/stats/get", h.GetClusterStats)
	g.POST("/kubernetes/events/list", h.ListEvents)
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
