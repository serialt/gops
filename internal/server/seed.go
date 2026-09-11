package server

import (
	"context"
	"log/slog"
	"strconv"

	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/data/model"
	"gops/internal/data/query"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

var (
	DEFAULT_PASSWORD = "password"
)

func seed(db *gorm.DB, casbinService *casbinx.Service) error {
	var admin model.User
	result := db.Where("username = ?", "admin").First(&admin)
	if result.Error != nil && result.Error != gorm.ErrRecordNotFound {
		return result.Error
	}
	if result.Error == gorm.ErrRecordNotFound {
		hash, err := auth.HashPassword(DEFAULT_PASSWORD)
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
	} else {
		slog.Info("系统已经初始化,跳过")
		return nil
	}

	adminRole := model.Role{Code: "admin", Name: "超级管理员", Description: "拥有全部后台接口权限"}
	if err := db.Where("code = ?", adminRole.Code).FirstOrCreate(&adminRole).Error; err != nil {
		return err
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

func ResetUserPassword(ctx context.Context, username string, pass string, cfg *config.Config) {
	if username == "" {
		slog.Info("用户名为空")
		return
	}

	db, cleanup, err := data.NewDB(cfg, slog.Default())
	if err != nil {
		slog.Info("connect db failed", "err", err)
		return
	}
	defer cleanup()
	mm := query.Use(db).User
	_, err = mm.WithContext(ctx).Where(mm.Username.Eq(username)).Find()
	if err != nil && err != gorm.ErrRecordNotFound {
		slog.Info("unknown err", "err", err)
		return
	}
	if err == gorm.ErrRecordNotFound {
		slog.Info("user not exists, reset failed")
		return
	}
	hash, err := auth.HashPassword(pass)
	if err != nil {
		slog.Info("gen hash password failed", "err", err)
	}
	_, err = mm.WithContext(ctx).Where(mm.Username.Eq(username)).UpdateSimple(mm.PasswordHash.Value(hash))
	if err != nil {
		slog.Info("reset user password failed", "err", err)
		return
	} else {
		slog.Info("reset user password succeed!!!")
	}
}
