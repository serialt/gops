package handler

import (
	"github.com/gin-gonic/gin"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
)

type AuditHandler struct{ biz *biz.AuditBiz }

func NewAuditHandler(b *biz.AuditBiz) *AuditHandler { return &AuditHandler{b} }

func (h *AuditHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/audit/operation-logs/list", h.ListOperationLogs)
	g.POST("/audit/operation-logs/delete", h.DeleteOperationLogs)
	g.POST("/audit/login-logs/list", h.ListLoginLogs)
	g.POST("/audit/login-logs/delete", h.DeleteLoginLogs)
	g.POST("/audit/data-logs/list", h.ListDataLogs)
}

// DeleteOperationLogs POST /api/v1/audit/operation-logs/delete 删除操作日志。
func (h *AuditHandler) DeleteOperationLogs(c *gin.Context) {
	r := new(api.AuditDeleteRequest)
	if auditBind(c, r) {
		if e := h.biz.DeleteOperations(r); e != nil {
			response.Error(c, 500, e.Error())
		} else {
			response.OK(c, nil)
		}
	}
}

func auditBind(c *gin.Context, r any) bool {
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid audit request")
		return false
	}
	return true
}

// DeleteLoginLogs POST /api/v1/audit/login-logs/delete 删除登录日志。
func (h *AuditHandler) DeleteLoginLogs(c *gin.Context) {
	r := new(api.AuditDeleteRequest)
	if auditBind(c, r) {
		if e := h.biz.DeleteLogins(r); e != nil {
			response.Error(c, 500, e.Error())
		} else {
			response.OK(c, nil)
		}
	}
}

// ListOperationLogs POST /api/v1/audit/operation-logs/list 查询操作日志。
func (h *AuditHandler) ListOperationLogs(c *gin.Context) {
	r := new(api.AuditListRequest)
	if auditBind(c, r) {
		d, e := h.biz.Operations(r)
		if e != nil {
			response.Error(c, 500, e.Error())
		} else {
			response.OK(c, d)
		}
	}
}

// ListLoginLogs POST /api/v1/audit/login-logs/list 查询登录日志。
func (h *AuditHandler) ListLoginLogs(c *gin.Context) {
	r := new(api.AuditListRequest)
	if auditBind(c, r) {
		d, e := h.biz.Logins(r)
		if e != nil {
			response.Error(c, 500, e.Error())
		} else {
			response.OK(c, d)
		}
	}
}

// ListDataLogs POST /api/v1/audit/data-logs/list 查询数据日志。
func (h *AuditHandler) ListDataLogs(c *gin.Context) {
	r := new(api.AuditListRequest)
	if auditBind(c, r) {
		d, e := h.biz.Data(r)
		if e != nil {
			response.Error(c, 500, e.Error())
		} else {
			response.OK(c, d)
		}
	}
}
