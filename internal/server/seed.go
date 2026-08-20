package server

import (
	"strconv"

	"gops/internal/data/model"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

func seed(db *gorm.DB, casbinService *casbinx.Service, adminPassword string) error {
	adminRole := model.Role{Code: "admin", Name: "超级管理员", Description: "拥有全部后台接口权限"}
	if err := db.Where("code = ?", adminRole.Code).FirstOrCreate(&adminRole).Error; err != nil {
		return err
	}

	var admin model.User
	result := db.Where("username = ?", "admin").First(&admin)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}
	if result.Error == gorm.ErrRecordNotFound {
		hash, err := auth.HashPassword(adminPassword)
		if err != nil {
			return err
		}
		admin = model.User{
			Username:     "admin",
			PasswordHash: hash,
			Nickname:     "管理员",
			Status:       true,
		}
		if err := db.Create(&admin).Error; err != nil {
			return err
		}
	}

	if err := db.Model(&admin).Association("Roles").Append(&adminRole); err != nil {
		return err
	}
	if _, err := casbinService.Enforcer.AddGroupingPolicy(strconv.FormatUint(uint64(admin.ID), 10), adminRole.Code); err != nil {
		return err
	}
	if _, err := casbinService.Enforcer.AddPolicy(adminRole.Code, "/api/v1/*", ".*"); err != nil {
		return err
	}
	return nil
}
