package handler

import (
	"context"
	"errors"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type MonitorHandler struct{ biz *biz.MonitorBiz }

func NewMonitorHandler(monitorBiz *biz.MonitorBiz) *MonitorHandler {
	return &MonitorHandler{biz: monitorBiz}
}

func (h *MonitorHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/monitor/domains/list", h.ListDomains)
	g.POST("/monitor/domains/create", h.CreateDomain)
	g.POST("/monitor/domains/update", h.UpdateDomain)
	g.POST("/monitor/domains/delete", h.DeleteDomain)
	g.POST("/monitor/domains/check", h.CheckDomain)
	g.POST("/monitor/domains/history", h.ListCheckHistory)
	g.POST("/monitor/domains/stats", h.DomainStats)

	g.POST("/monitor/channels/list", h.ListChannels)
	g.POST("/monitor/channels/save", h.SaveChannel)
	g.POST("/monitor/channels/delete", h.DeleteChannel)
	g.POST("/monitor/channels/test", h.TestChannel)

	g.POST("/monitor/receivers/list", h.ListReceivers)
	g.POST("/monitor/receivers/save", h.SaveReceiver)
	g.POST("/monitor/receivers/delete", h.DeleteReceiver)

	g.POST("/monitor/logs/list", h.ListLogs)
}

func (h *MonitorHandler) StartScheduler(ctx context.Context) {
	h.biz.StartScheduler(ctx)
}

func monitorBind(c *gin.Context, req any) bool {
	if c.ShouldBindJSON(req) != nil {
		response.Error(c, 400, "invalid monitor request")
		return false
	}
	return true
}
func monitorResult(c *gin.Context, data any, err error) {
	if err == nil {
		response.OK(c, data)
		return
	}
	if errors.Is(err, biz.ErrNotFound) {
		response.Error(c, 404, "monitor resource not found")
		return
	}
	response.Error(c, 400, err.Error())
}

// ListDomains POST /api/v1/monitor/domains/list 查询域名监控。
func (h *MonitorHandler) ListDomains(c *gin.Context) {
	req := new(api.MonitorDomainListRequest)
	if monitorBind(c, req) {
		data, err := h.biz.ListDomains(req)
		monitorResult(c, data, err)
	}
}

// CreateDomain POST /api/v1/monitor/domains/create 创建域名监控。
func (h *MonitorHandler) CreateDomain(c *gin.Context) {
	req := new(api.MonitorDomainRequest)
	if monitorBind(c, req) {
		data, err := h.biz.CreateDomain(req)
		monitorResult(c, data, err)
	}
}

// UpdateDomain POST /api/v1/monitor/domains/update 更新域名监控。
func (h *MonitorHandler) UpdateDomain(c *gin.Context) {
	req := new(api.MonitorDomainRequest)
	if monitorBind(c, req) {
		data, err := h.biz.UpdateDomain(req)
		monitorResult(c, data, err)
	}
}

// DeleteDomain POST /api/v1/monitor/domains/delete 删除域名监控。
func (h *MonitorHandler) DeleteDomain(c *gin.Context) {
	req := new(api.MonitorDomainIDRequest)
	if monitorBind(c, req) {
		monitorResult(c, nil, h.biz.DeleteDomain(req))
	}
}

// CheckDomain POST /api/v1/monitor/domains/check 立即检测域名。
func (h *MonitorHandler) CheckDomain(c *gin.Context) {
	req := new(api.MonitorDomainIDRequest)
	if monitorBind(c, req) {
		data, err := h.biz.CheckDomain(req)
		monitorResult(c, data, err)
	}
}

// ListCheckHistory POST /api/v1/monitor/domains/history 查询域名请求日志。
func (h *MonitorHandler) ListCheckHistory(c *gin.Context) {
	req := new(api.MonitorCheckHistoryListRequest)
	if monitorBind(c, req) {
		data, err := h.biz.ListCheckHistory(req)
		monitorResult(c, data, err)
	}
}

// DomainStats POST /api/v1/monitor/domains/stats 查询监控统计。
func (h *MonitorHandler) DomainStats(c *gin.Context) {
	req := new(struct{})
	if monitorBind(c, req) {
		data, err := h.biz.DomainStats()
		monitorResult(c, data, err)
	}
}

// ListChannels POST /api/v1/monitor/channels/list 查询告警通道。
func (h *MonitorHandler) ListChannels(c *gin.Context) {
	req := new(api.MonitorAlertChannelListRequest)
	if monitorBind(c, req) {
		data, err := h.biz.ListChannels(req)
		monitorResult(c, data, err)
	}
}

// SaveChannel POST /api/v1/monitor/channels/save 保存告警通道。
func (h *MonitorHandler) SaveChannel(c *gin.Context) {
	req := new(api.MonitorAlertChannelRequest)
	if monitorBind(c, req) {
		data, err := h.biz.SaveChannel(req)
		monitorResult(c, data, err)
	}
}

// DeleteChannel POST /api/v1/monitor/channels/delete 删除告警通道。
func (h *MonitorHandler) DeleteChannel(c *gin.Context) {
	req := new(api.MonitorAlertChannelIDRequest)
	if monitorBind(c, req) {
		monitorResult(c, nil, h.biz.DeleteChannel(req))
	}
}

// TestChannel POST /api/v1/monitor/channels/test 测试告警通道。
func (h *MonitorHandler) TestChannel(c *gin.Context) {
	req := new(api.MonitorAlertChannelTestRequest)
	if monitorBind(c, req) {
		monitorResult(c, nil, h.biz.TestChannel(req))
	}
}

// ListReceivers POST /api/v1/monitor/receivers/list 查询告警接收人。
func (h *MonitorHandler) ListReceivers(c *gin.Context) {
	req := new(api.MonitorAlertReceiverListRequest)
	if monitorBind(c, req) {
		data, err := h.biz.ListReceivers(req)
		monitorResult(c, data, err)
	}
}

// SaveReceiver POST /api/v1/monitor/receivers/save 保存告警接收人。
func (h *MonitorHandler) SaveReceiver(c *gin.Context) {
	req := new(api.MonitorAlertReceiverRequest)
	if monitorBind(c, req) {
		data, err := h.biz.SaveReceiver(req)
		monitorResult(c, data, err)
	}
}

// DeleteReceiver POST /api/v1/monitor/receivers/delete 删除告警接收人。
func (h *MonitorHandler) DeleteReceiver(c *gin.Context) {
	req := new(api.MonitorAlertReceiverIDRequest)
	if monitorBind(c, req) {
		monitorResult(c, nil, h.biz.DeleteReceiver(req))
	}
}

// ListLogs POST /api/v1/monitor/logs/list 查询告警日志。
func (h *MonitorHandler) ListLogs(c *gin.Context) {
	req := new(api.MonitorAlertLogListRequest)
	if monitorBind(c, req) {
		data, err := h.biz.ListLogs(req)
		monitorResult(c, data, err)
	}
}
