package handler

import (
	"errors"
	"strconv"

	"gops/internal/data/model"
	"gops/internal/pkg/casbinx"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db     *gorm.DB
	casbin *casbinx.Service
}

func NewRoleHandler(db *gorm.DB, casbinService *casbinx.Service) *RoleHandler {
	return &RoleHandler{db: db, casbin: casbinService}
}

type roleRequest struct {
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type permissionRequest struct {
	Permissions []casbinx.Permission `json:"permissions"`
}

func (h *RoleHandler) List(c *gin.Context) {
	var roles []model.Role
	if err := h.db.Order("id ASC").Find(&roles).Error; err != nil {
		response.Error(c, 500, "query roles failed")
		return
	}
	response.OK(c, roles)
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid role data")
		return
	}
	role := model.Role{Code: req.Code, Name: req.Name, Description: req.Description}
	if err := h.db.Create(&role).Error; err != nil {
		response.Error(c, 400, "create role failed: "+err.Error())
		return
	}
	response.Created(c, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "invalid role id")
		return
	}
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid role data")
		return
	}
	var role model.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		response.Error(c, 404, "role not found")
		return
	}
	if req.Code != role.Code {
		response.Error(c, 400, "role code cannot be changed")
		return
	}
	role.Name = req.Name
	role.Description = req.Description
	if err := h.db.Save(&role).Error; err != nil {
		response.Error(c, 400, "update role failed")
		return
	}
	response.OK(c, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "invalid role id")
		return
	}
	var role model.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		response.Error(c, 404, "role not found")
		return
	}
	if role.Code == "admin" {
		response.Error(c, 400, "admin role cannot be deleted")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", role.ID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	}); err != nil {
		response.Error(c, 500, "delete role failed")
		return
	}
	if err := h.casbin.DeleteRole(role.Code); err != nil {
		response.Error(c, 500, "delete role policy failed")
		return
	}
	response.OK(c, nil)
}

func (h *RoleHandler) Catalog(c *gin.Context) {
	response.OK(c, casbinx.Catalog())
}

func (h *RoleHandler) GetPermissions(c *gin.Context) {
	role, err := h.findRole(c.Param("id"))
	if err != nil {
		response.Error(c, 404, "role not found")
		return
	}

	policies := h.casbin.GetRolePolicies(role.Code)
	permissions := make([]casbinx.Permission, 0, len(policies))
	for _, policy := range policies {
		if len(policy) < 3 {
			continue
		}
		permissions = append(permissions, casbinx.Permission{
			Path: policy[1], Method: policy[2], Key: policy[2] + " " + policy[1],
		})
	}
	response.OK(c, permissions)
}

func (h *RoleHandler) SetPermissions(c *gin.Context) {
	role, err := h.findRole(c.Param("id"))
	if err != nil {
		response.Error(c, 404, "role not found")
		return
	}
	if role.Code == "admin" {
		response.Error(c, 400, "admin role keeps full access")
		return
	}
	var req permissionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid permission data")
		return
	}

	catalog := make(map[string]casbinx.Permission)
	for _, item := range casbinx.Catalog() {
		catalog[item.Key] = item
	}
	clean := make([]casbinx.Permission, 0, len(req.Permissions))
	seen := make(map[string]bool)
	for _, item := range req.Permissions {
		key := item.Method + " " + item.Path
		value, ok := catalog[key]
		if !ok {
			response.Error(c, 400, "unknown permission: "+key)
			return
		}
		if !seen[key] {
			clean = append(clean, value)
			seen[key] = true
		}
	}

	if err := h.casbin.SetRolePolicies(role.Code, clean); err != nil {
		response.Error(c, 500, "save permissions failed")
		return
	}
	response.OK(c, clean)
}

func (h *RoleHandler) findRole(value string) (*model.Role, error) {
	id, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return nil, err
	}
	var role model.Role
	if err := h.db.First(&role, uint(id)).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err
		}
		return nil, err
	}
	return &role, nil
}
