package handler

import (
	"gops/internal/data/model"
	"gops/internal/middleware"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db     *gorm.DB
	jwt    *auth.JWTManager
	casbin *casbinx.Service
}

func NewAuthHandler(db *gorm.DB, jwt *auth.JWTManager, casbinService *casbinx.Service) *AuthHandler {
	return &AuthHandler{db: db, jwt: jwt, casbin: casbinService}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "username and password are required")
		return
	}

	var user model.User
	if err := h.db.Preload("Roles").Where("username = ?", req.Username).First(&user).Error; err != nil {
		response.Error(c, 401, "invalid username or password")
		return
	}
	if !user.Status || !auth.CheckPassword(user.PasswordHash, req.Password) {
		response.Error(c, 401, "invalid username or password")
		return
	}

	token, err := h.jwt.Generate(uint(user.ID))
	if err != nil {
		response.Error(c, 500, "generate token failed")
		return
	}
	response.OK(c, gin.H{"token": token})
}

func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}

	var user model.User
	if err := h.db.Preload("Roles").First(&user, userID).Error; err != nil {
		response.Error(c, 404, "user not found")
		return
	}

	response.OK(c, gin.H{
		"user":        user,
		"permissions": h.casbin.AllowedPermissionKeys(uint(user.ID)),
	})
}
