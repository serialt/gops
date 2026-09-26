package handler

import (
	"errors"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type RoleHandler struct {
	biz *biz.RoleBiz
}

func NewRoleHandler(roleBiz *biz.RoleBiz) *RoleHandler {
	return &RoleHandler{biz: roleBiz}
}

func (h *RoleHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/roles/list", h.List)
	g.POST("/roles/create", h.Create)
	g.POST("/roles/update", h.Update)
	g.POST("/roles/delete", h.Delete)
	g.POST("/roles/permissions/get", h.GetPermissions)
	g.POST("/roles/permissions/set", h.SetPermissions)
	g.POST("/permissions/catalog", h.Catalog)
}

// List POST /api/v1/roles/list 查询角色列表。
func (h *RoleHandler) List(c *gin.Context) {
	req := new(api.RoleListRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid role query")
		return
	}
	result, err := h.biz.List(req)
	if err != nil {
		response.Error(c, 500, "query roles failed")
		return
	}
	response.OK(c, result)
}

// Create POST /api/v1/roles/create 创建角色。
func (h *RoleHandler) Create(c *gin.Context) {
	req := new(api.RoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid role data")
		return
	}
	result, err := h.biz.Create(req)
	if err != nil {
		response.Error(c, 400, "create role failed: "+err.Error())
		return
	}
	response.Created(c, result)
}

// Update POST /api/v1/roles/update 更新角色。
func (h *RoleHandler) Update(c *gin.Context) {
	req := new(api.RoleRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid role data")
		return
	}
	result, err := h.biz.Update(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "role not found")
			return
		}
		if errors.Is(err, biz.ErrInvalid) {
			response.Error(c, 400, "invalid role id")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// Delete POST /api/v1/roles/delete 删除角色。
func (h *RoleHandler) Delete(c *gin.Context) {
	req := new(api.RoleIDRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid role id")
		return
	}
	if err := h.biz.Delete(req); err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "role not found")
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}

// Catalog POST /api/v1/permissions/catalog 查询权限目录。
func (h *RoleHandler) Catalog(c *gin.Context) {
	response.OK(c, h.biz.Catalog())
}

// GetPermissions POST /api/v1/roles/permissions/get 查询角色权限。
func (h *RoleHandler) GetPermissions(c *gin.Context) {
	req := new(api.PermissionRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid role id")
		return
	}
	result, err := h.biz.GetPermissions(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "role not found")
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}

// SetPermissions POST /api/v1/roles/permissions/set 保存角色权限。
func (h *RoleHandler) SetPermissions(c *gin.Context) {
	req := new(api.PermissionRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid permission data")
		return
	}
	result, err := h.biz.SetPermissions(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "role not found")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}
