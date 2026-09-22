package handler

import (
	"gops/internal/biz"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	biz *biz.DashboardBiz
}

func NewDashboardHandler(dashboardBiz *biz.DashboardBiz) *DashboardHandler {
	return &DashboardHandler{biz: dashboardBiz}
}

func (h *DashboardHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/dashboard/stats", h.Stats)
}

// Stats POST /api/v1/dashboard/stats 获取仪表盘统计数据。
func (h *DashboardHandler) Stats(c *gin.Context) {
	response.OK(c, h.biz.Stats())
}
