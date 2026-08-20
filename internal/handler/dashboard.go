package handler

import (
	"gops/internal/data/model"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DashboardHandler struct {
	db *gorm.DB
}

func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{db: db}
}

func (h *DashboardHandler) Stats(c *gin.Context) {
	var users, enabledUsers, roles, policies int64
	h.db.Model(&model.User{}).Count(&users)
	h.db.Model(&model.User{}).Where("status = ?", true).Count(&enabledUsers)
	h.db.Model(&model.Role{}).Count(&roles)
	h.db.Table("casbin_rule").Where("ptype = ?", "p").Count(&policies)
	response.OK(c, gin.H{
		"users":         users,
		"enabled_users": enabledUsers,
		"roles":         roles,
		"policies":      policies,
	})
}
