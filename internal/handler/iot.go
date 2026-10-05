package handler

import (
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type IOTHandler struct{ biz *biz.IOTBiz }

func NewIOTHandler(b *biz.IOTBiz) *IOTHandler { return &IOTHandler{biz: b} }
func (h *IOTHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/iot/dashboard/stats", h.Dashboard)
	g.POST("/iot/acl-rules/list", h.ListACLRules)
	g.POST("/iot/acl-rules/create", h.CreateACLRule)
	g.POST("/iot/acl-rules/update", h.UpdateACLRule)
	g.POST("/iot/acl-rules/delete", h.DeleteACLRule)
	g.POST("/iot/topics/list", h.ListTopics)
	g.POST("/iot/messages/list", h.ListMessages)
	g.POST("/iot/clients/list", h.ListClients)
	g.POST("/iot/messages/publish", h.Publish)
}

// Dashboard POST /api/v1/iot/dashboard/stats 查询 IOT 大屏聚合数据。
func (h *IOTHandler) Dashboard(c *gin.Context) {
	data, err := h.biz.Dashboard(c.Request.Context())
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}
func bind[T any](c *gin.Context) (*T, bool) {
	req := new(T)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请求参数错误")
		return nil, false
	}
	return req, true
}
func handleErr(c *gin.Context, err error) {
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// ListACLRules POST /api/v1/iot/acl-rules/list 查询 IOT ACL 规则。
func (h *IOTHandler) ListACLRules(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListACLRules(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// CreateACLRule POST /api/v1/iot/acl-rules/create 新建 IOT ACL 规则。
func (h *IOTHandler) CreateACLRule(c *gin.Context) {
	req, ok := bind[api.IOTACLRuleRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.CreateACLRule(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// UpdateACLRule POST /api/v1/iot/acl-rules/update 更新 IOT ACL 规则。
func (h *IOTHandler) UpdateACLRule(c *gin.Context) {
	req, ok := bind[api.IOTACLRuleRequest](c)
	if ok {
		handleErr(c, h.biz.UpdateACLRule(c.Request.Context(), req))
	}
}

// DeleteACLRule POST /api/v1/iot/acl-rules/delete 删除 IOT ACL 规则。
func (h *IOTHandler) DeleteACLRule(c *gin.Context) {
	req, ok := bind[api.IOTIDRequest](c)
	if ok {
		handleErr(c, h.biz.DeleteACLRule(c.Request.Context(), req))
	}
}

// ListTopics POST /api/v1/iot/topics/list 查询 IOT Topic。
func (h *IOTHandler) ListTopics(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListTopics(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// ListMessages POST /api/v1/iot/messages/list 查询 IOT 接收消息。
func (h *IOTHandler) ListMessages(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListMessages(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// ListClients POST /api/v1/iot/clients/list 查询 IOT 客户端记录。
func (h *IOTHandler) ListClients(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListClients(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// Publish POST /api/v1/iot/messages/publish 发布 IOT MQTT 消息。
func (h *IOTHandler) Publish(c *gin.Context) {
	req, ok := bind[api.IOTPublishRequest](c)
	if ok {
		handleErr(c, h.biz.Publish(c.Request.Context(), req))
	}
}
