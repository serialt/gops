package handler

import (
	"errors"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type KubernetesHandler struct{ biz *biz.KubernetesBiz }

func NewKubernetesHandler(b *biz.KubernetesBiz) *KubernetesHandler { return &KubernetesHandler{b} }

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
