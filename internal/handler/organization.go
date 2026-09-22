package handler

import (
	"errors"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type OrganizationHandler struct {
	biz *biz.OrganizationBiz
}

func NewOrganizationHandler(organizationBiz *biz.OrganizationBiz) *OrganizationHandler {
	return &OrganizationHandler{biz: organizationBiz}
}

func (h *OrganizationHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/organizations/list", h.List)
	g.POST("/organizations/create", h.Create)
	g.POST("/organizations/update", h.Update)
	g.POST("/organizations/delete", h.Delete)
	g.POST("/organizations/move", h.Move)
}
// List POST /api/v1/organizations/list 查询组织列表。
func (h *OrganizationHandler) List(c *gin.Context) {
	req := new(api.OrganizationQueryRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid organization query")
		return
	}
	result, err := h.biz.List(req)
	if err != nil {
		response.Error(c, 500, "query organizations failed")
		return
	}
	response.OK(c, result)
}

// Create POST /api/v1/organizations/create 创建组织。
func (h *OrganizationHandler) Create(c *gin.Context) {
	req := new(api.OrganizationRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid organization data")
		return
	}
	result, err := h.biz.Create(req)
	if err != nil {
		response.Error(c, 400, "create organization failed: "+err.Error())
		return
	}
	response.Created(c, result)
}

// Update POST /api/v1/organizations/update 更新组织。
func (h *OrganizationHandler) Update(c *gin.Context) {
	req := new(api.OrganizationRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid organization data")
		return
	}
	result, err := h.biz.Update(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "organization not found")
			return
		}
		if errors.Is(err, biz.ErrInvalid) {
			response.Error(c, 400, "invalid organization id")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// Delete POST /api/v1/organizations/delete 删除组织。
func (h *OrganizationHandler) Delete(c *gin.Context) {
	req := new(api.OrganizationIDRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid organization id")
		return
	}
	if err := h.biz.Delete(req); err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "organization not found")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// Move POST /api/v1/organizations/move 移动并排序组织。
func (h *OrganizationHandler) Move(c *gin.Context) {
	req := new(api.OrganizationMoveRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid move data")
		return
	}
	result, err := h.biz.Move(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "organization not found")
			return
		}
		response.Error(c, 500, "move organization failed")
		return
	}
	response.OK(c, result)
}
