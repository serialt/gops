package biz

import (
	"errors"

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

func (b *AuthBiz) Login(req *api.LoginRequest) (*api.LoginResponse, error) {
	var user model.User
	if err := b.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	if !user.Status || !auth.CheckPassword(user.PasswordHash, req.Password) {
		return nil, ErrInvalidCredentials
	}

	token, err := b.jwt.Generate(uint(user.ID))
	if err != nil {
		return nil, ErrGenerateToken
	}
	return &api.LoginResponse{Token: token}, nil
}

func (b *AuthBiz) Me(userID uint) (*api.CurrentUserResponse, error) {
	var user model.User
	if err := b.db.Preload("Roles").Preload("Organizations").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	return &api.CurrentUserResponse{User: toUserResponse(user), Permissions: b.casbin.AllowedPermissionKeys(userID)}, nil
}
