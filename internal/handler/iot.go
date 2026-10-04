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
	g.POST("/iot/users/list", h.ListUsers)
	g.POST("/iot/users/create", h.CreateUser)
	g.POST("/iot/users/update", h.UpdateUser)
	g.POST("/iot/users/delete", h.DeleteUser)
	g.POST("/iot/roles/list", h.ListRoles)
	g.POST("/iot/roles/create", h.CreateRole)
	g.POST("/iot/roles/update", h.UpdateRole)
	g.POST("/iot/roles/delete", h.DeleteRole)
	g.POST("/iot/acl-rules/list", h.ListACLRules)
	g.POST("/iot/acl-rules/create", h.CreateACLRule)
	g.POST("/iot/acl-rules/update", h.UpdateACLRule)
	g.POST("/iot/acl-rules/delete", h.DeleteACLRule)
	g.POST("/iot/topics/list", h.ListTopics)
	g.POST("/iot/messages/list", h.ListMessages)
	g.POST("/iot/clients/list", h.ListClients)
	g.POST("/iot/messages/publish", h.Publish)
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

// ListUsers POST /api/v1/iot/users/list 查询 IOT 用户。
func (h *IOTHandler) ListUsers(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListUsers(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// CreateUser POST /api/v1/iot/users/create 新建 IOT 用户。
func (h *IOTHandler) CreateUser(c *gin.Context) {
	req, ok := bind[api.IOTUserRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.CreateUser(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// UpdateUser POST /api/v1/iot/users/update 更新 IOT 用户。
func (h *IOTHandler) UpdateUser(c *gin.Context) {
	req, ok := bind[api.IOTUserRequest](c)
	if ok {
		handleErr(c, h.biz.UpdateUser(c.Request.Context(), req))
	}
}

// DeleteUser POST /api/v1/iot/users/delete 删除 IOT 用户。
func (h *IOTHandler) DeleteUser(c *gin.Context) {
	req, ok := bind[api.IOTIDRequest](c)
	if ok {
		handleErr(c, h.biz.DeleteUser(c.Request.Context(), req))
	}
}

// ListRoles POST /api/v1/iot/roles/list 查询 IOT 角色。
func (h *IOTHandler) ListRoles(c *gin.Context) {
	req, ok := bind[api.IOTListRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.ListRoles(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, data)
}

// CreateRole POST /api/v1/iot/roles/create 新建 IOT 角色。
func (h *IOTHandler) CreateRole(c *gin.Context) {
	req, ok := bind[api.IOTRoleRequest](c)
	if !ok {
		return
	}
	data, err := h.biz.CreateRole(c.Request.Context(), req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, data)
}

// UpdateRole POST /api/v1/iot/roles/update 更新 IOT 角色。
func (h *IOTHandler) UpdateRole(c *gin.Context) {
	req, ok := bind[api.IOTRoleRequest](c)
	if ok {
		handleErr(c, h.biz.UpdateRole(c.Request.Context(), req))
	}
}

// DeleteRole POST /api/v1/iot/roles/delete 删除 IOT 角色。
func (h *IOTHandler) DeleteRole(c *gin.Context) {
	req, ok := bind[api.IOTIDRequest](c)
	if ok {
		handleErr(c, h.biz.DeleteRole(c.Request.Context(), req))
	}
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
