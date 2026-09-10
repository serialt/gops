package biz

import (
	"gops/internal/data/api"
	"gops/internal/data/model"

	"gorm.io/gorm"
)

type DashboardBiz struct {
	db *gorm.DB
}

func NewDashboardBiz(db *gorm.DB) *DashboardBiz {
	return &DashboardBiz{db: db}
}

func (b *DashboardBiz) Stats() *api.DashboardStatsResponse {
	result := &api.DashboardStatsResponse{}
	b.db.Model(&model.User{}).Count(&result.Users)
	b.db.Model(&model.User{}).Where("status = ?", true).Count(&result.EnabledUsers)
	b.db.Model(&model.Role{}).Count(&result.Roles)
	b.db.Table("casbin_rule").Where("ptype = ?", "p").Count(&result.Policies)
	return result
}
