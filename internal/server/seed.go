package server

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strconv"

	"gops/internal/biz"
	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/data/model"
	"gops/internal/data/repo/rbac"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

var (
	DEFAULT_PASSWORD = "password"
)

// Seed 初始化内置管理员、管理员角色及其完整接口权限。
func Seed(db *gorm.DB, casbinService *casbinx.Service) error {
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
		slog.Info("系统已经初始化，校验管理员权限")
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

// ManageUsers 执行命令行用户列表、创建和密码重置操作。
func ManageUsers(ctx context.Context, action string, args []string, cfg *config.Config) error {
	db, cleanup, err := data.NewDB(cfg, slog.Default())
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer cleanup()
	casbinService, err := casbinx.NewService(db)
	if err != nil {
		return fmt.Errorf("initialize Casbin: %w", err)
	}
	userBiz := biz.NewUserBiz(db, casbinService, rbac.NewRbacRepo())

	switch action {
	case "list":
		if len(args) != 0 {
			return errors.New("usage: -u list")
		}
		users, err := userBiz.ListForCLI(ctx)
		if err != nil {
			return err
		}
		fmt.Printf("%-6s %-24s %-24s %-8s\n", "ID", "USERNAME", "NICKNAME", "STATUS")
		for _, user := range users {
			status := "disabled"
			if user.Status {
				status = "enabled"
			}
			fmt.Printf("%-6d %-24s %-24s %-8s\n", user.ID, user.Username, user.Nickname, status)
		}
		return nil
	case "create":
		if len(args) < 2 || len(args) > 3 {
			return errors.New("usage: -u create <username> <password> [nickname]")
		}
		nickname := args[0]
		if len(args) == 3 {
			nickname = args[2]
		}
		user, err := userBiz.CreateForCLI(ctx, args[0], args[1], nickname)
		if err != nil {
			return err
		}
		fmt.Printf("created administrator %s (ID %d)\n", user.Username, user.ID)
		return nil
	case "reset":
		if len(args) != 2 {
			return errors.New("usage: -u reset <username> <new-password>")
		}
		if err := userBiz.ResetPasswordForCLI(ctx, args[0], args[1]); err != nil {
			return err
		}
		fmt.Printf("reset password for user %s\n", args[0])
		return nil
	default:
		if len(args) == 1 {
			if err := userBiz.ResetPasswordForCLI(ctx, action, args[0]); err != nil {
				return err
			}
			fmt.Printf("reset password for user %s\n", action)
			return nil
		}
		return errors.New("unknown user action; use list, create, or reset")
	}
}
