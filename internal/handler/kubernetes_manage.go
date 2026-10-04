package handler

import (
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// ListKubernetesObjects POST /api/v1/kubernetes/resources/list 查询 Kubernetes 白名单资源。
func (h *KubernetesHandler) ListKubernetesObjects(c *gin.Context) {
	r := new(api.KubernetesObjectRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource query")
		return
	}
	data, err := h.biz.ListObjects(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetKubernetesObject POST /api/v1/kubernetes/resources/get 获取 Kubernetes 资源 YAML。
func (h *KubernetesHandler) GetKubernetesObject(c *gin.Context) {
	r := new(api.KubernetesObjectRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource query")
		return
	}
	data, err := h.biz.GetObjectYAML(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetKubernetesStorageResourceSupport POST /api/v1/kubernetes/storage/resources/support 查询 PV、PVC 或 StorageClass 详情。
func (h *KubernetesHandler) GetKubernetesStorageResourceSupport(c *gin.Context) {
	req := new(api.KubernetesStorageDetailRequest)
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid storage resource query")
		return
	}
	data, err := h.biz.StorageResourceSupport(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetKubernetesResourceTemplate POST /api/v1/kubernetes/resources/template 获取 Kubernetes 资源创建模板。
func (h *KubernetesHandler) GetKubernetesResourceTemplate(c *gin.Context) {
	r := new(api.KubernetesObjectRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource template query")
		return
	}
	data, err := h.biz.ResourceTemplate(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ApplyKubernetesObject POST /api/v1/kubernetes/resources/apply 创建或更新 Kubernetes 资源。
func (h *KubernetesHandler) ApplyKubernetesObject(c *gin.Context) {
	r := new(api.KubernetesObjectRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource data")
		return
	}
	data, err := h.biz.ApplyObject(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// DeleteKubernetesObject POST /api/v1/kubernetes/resources/delete 删除 Kubernetes 资源。
func (h *KubernetesHandler) DeleteKubernetesObject(c *gin.Context) {
	r := new(api.KubernetesObjectRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource data")
		return
	}
	if err := h.biz.DeleteObject(r); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// UpdateKubernetesObjectLabels POST /api/v1/kubernetes/resources/labels/update 更新 Kubernetes 资源标签。
func (h *KubernetesHandler) UpdateKubernetesObjectLabels(c *gin.Context) {
	r := new(api.KubernetesLabelUpdateRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid resource labels")
		return
	}
	data, err := h.biz.UpdateObjectLabels(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}
