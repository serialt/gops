package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"
)

type NginxHandler struct{ biz *biz.NginxBiz }

func NewNginxHandler(b *biz.NginxBiz) *NginxHandler { return &NginxHandler{b} }

func (h *NginxHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/nginx/sources/list", h.ListSources)
	g.POST("/nginx/sources/save", h.SaveSource)
	g.POST("/nginx/sources/delete", h.DeleteSource)
	g.POST("/nginx/overview", h.Overview)
	g.POST("/nginx/logs/list", h.ListLogs)
}
func nginxBind(c *gin.Context, r any) bool {
	if c.ShouldBindJSON(r) != nil {
		response.Error(c, 400, "invalid nginx request")
		return false
	}
	return true
}
func nginxReply(c *gin.Context, d any, e error) {
	if e == nil {
		response.OK(c, d)
	} else if errors.Is(e, biz.ErrNotFound) {
		response.Error(c, 404, "nginx resource not found")
	} else {
		response.Error(c, 400, e.Error())
	}
}

// ListSources POST /api/v1/nginx/sources/list 查询 Nginx 数据源。
func (h *NginxHandler) ListSources(c *gin.Context) {
	r := new(api.NginxSourceListRequest)
	if nginxBind(c, r) {
		d, e := h.biz.ListSources(r)
		nginxReply(c, d, e)
	}
}

// SaveSource POST /api/v1/nginx/sources/save 保存 Nginx 数据源。
func (h *NginxHandler) SaveSource(c *gin.Context) {
	r := new(api.NginxSourceRequest)
	if nginxBind(c, r) {
		d, e := h.biz.SaveSource(r)
		nginxReply(c, d, e)
	}
}

// DeleteSource POST /api/v1/nginx/sources/delete 删除 Nginx 数据源。
func (h *NginxHandler) DeleteSource(c *gin.Context) {
	r := new(api.NginxSourceIDRequest)
	if nginxBind(c, r) {
		nginxReply(c, nil, h.biz.DeleteSource(r))
	}
}

// Overview POST /api/v1/nginx/overview 查询 Nginx 概况。
func (h *NginxHandler) Overview(c *gin.Context) {
	r := new(struct{})
	if nginxBind(c, r) {
		d, e := h.biz.Overview()
		nginxReply(c, d, e)
	}
}

// ListLogs POST /api/v1/nginx/logs/list 查询 Nginx 访问明细。
func (h *NginxHandler) ListLogs(c *gin.Context) {
	r := new(api.NginxAccessLogListRequest)
	if nginxBind(c, r) {
		d, e := h.biz.ListLogs(r)
		nginxReply(c, d, e)
	}
}
