package handler

import (
	"gops/internal/data/api"
	"gops/internal/middleware"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

// StartInspection POST /api/v1/kubernetes/inspections/start 启动集群巡检。
func (h *KubernetesHandler) StartInspection(c *gin.Context) {
	r := new(api.KubernetesInspectionStartRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid inspection request")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	r.UserID = int64(userID)
	data, err := h.biz.StartInspection(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetInspectionProgress POST /api/v1/kubernetes/inspections/progress 查询集群巡检进度。
func (h *KubernetesHandler) GetInspectionProgress(c *gin.Context) {
	r := new(api.KubernetesInspectionIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid inspection id")
		return
	}
	data, err := h.biz.InspectionProgress(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// GetInspectionResult POST /api/v1/kubernetes/inspections/result 查询集群巡检报告。
func (h *KubernetesHandler) GetInspectionResult(c *gin.Context) {
	r := new(api.KubernetesInspectionIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid inspection id")
		return
	}
	data, err := h.biz.InspectionResult(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// ListInspectionHistory POST /api/v1/kubernetes/inspections/history 查询集群巡检历史。
func (h *KubernetesHandler) ListInspectionHistory(c *gin.Context) {
	r := new(api.KubernetesInspectionHistoryRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid inspection query")
		return
	}
	data, err := h.biz.InspectionHistory(r)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// DeleteInspection POST /api/v1/kubernetes/inspections/delete 删除集群巡检记录。
func (h *KubernetesHandler) DeleteInspection(c *gin.Context) {
	r := new(api.KubernetesInspectionIDRequest)
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid inspection id")
		return
	}
	if err := h.biz.DeleteInspection(r); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}
