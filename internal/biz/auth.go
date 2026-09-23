package biz

import (
	"errors"
	"log/slog"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrGenerateToken      = errors.New("generate token failed")
)

type AuthBiz struct {
	db     *gorm.DB
	jwt    *auth.JWTManager
	casbin *casbinx.Service
}

func NewAuthBiz(db *gorm.DB, jwt *auth.JWTManager, casbinService *casbinx.Service) *AuthBiz {
	return &AuthBiz{db: db, jwt: jwt, casbin: casbinService}
}

func (b *AuthBiz) Login(req *api.LoginRequest, ip string) (*api.LoginResponse, error) {
	var user model.User
	if err := b.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		b.recordLogin(0, req.Username, ip, "failed", "用户名或密码错误")
		return nil, ErrInvalidCredentials
	}
	if !user.Status || !auth.CheckPassword(user.PasswordHash, req.Password) {
		message := "用户名或密码错误"
		if !user.Status {
			message = "用户已被禁用"
		}
		b.recordLogin(user.ID, user.Username, ip, "failed", message)
		return nil, ErrInvalidCredentials
	}

	token, err := b.jwt.Generate(uint(user.ID))
	if err != nil {
		b.recordLogin(user.ID, user.Username, ip, "failed", "生成登录令牌失败")
		return nil, ErrGenerateToken
	}
	b.recordLogin(user.ID, user.Username, ip, "success", "登录成功")
	return &api.LoginResponse{Token: token}, nil
}

func (b *AuthBiz) RecordInvalidLogin(username, ip, message string) {
	b.recordLogin(0, username, ip, "failed", message)
}

func (b *AuthBiz) recordLogin(userID int64, username, ip, status, message string) {
	log := &model.AuditLoginLog{UserID: userID, Username: username, LoginType: "web", Status: status, IP: ip, Message: message}
	if err := b.db.Create(log).Error; err != nil {
		slog.Error("保存登录日志失败", "error", err, "username", username, "ip", ip)
	}
}

func (b *AuthBiz) Me(userID uint) (*api.CurrentUserResponse, error) {
	var user model.User
	if err := b.db.Preload("Roles").Preload("Organizations").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	return &api.CurrentUserResponse{User: toUserResponse(user), Permissions: b.casbin.AllowedPermissionKeys(userID)}, nil
}
