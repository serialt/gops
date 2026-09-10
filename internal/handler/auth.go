package handler

import (
	"errors"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/middleware"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	biz *biz.AuthBiz
}

func NewAuthHandler(authBiz *biz.AuthBiz) *AuthHandler {
	return &AuthHandler{biz: authBiz}
}

// Login POST /api/v1/auth/login 用户登录。
func (h *AuthHandler) Login(c *gin.Context) {
	req := new(api.LoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "username and password are required")
		return
	}

	result, err := h.biz.Login(req)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidCredentials) {
			response.Error(c, 401, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}

// Me POST /api/v1/auth/me 获取当前登录用户及权限。
func (h *AuthHandler) Me(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}

	result, err := h.biz.Me(userID)
	if err != nil {
		response.Error(c, 404, "user not found")
		return
	}
	response.OK(c, result)
}
