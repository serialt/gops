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

func (h *AuthHandler) RegisterRoutes(g *gin.RouterGroup) {
	g.POST("/auth/me", h.Me)
	g.POST("/auth/profile/update", h.UpdateProfile)
}

// UpdateProfile POST /api/v1/auth/profile/update 更新当前用户的个性化资料。
func (h *AuthHandler) UpdateProfile(c *gin.Context) {
	req := new(api.ProfileUpdateRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "用户名、昵称或邮箱格式不正确")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	result, err := h.biz.UpdateProfile(userID, req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

func (h *AuthHandler) RegisterRoutesNologin(g *gin.RouterGroup) {
	g.POST("/auth/login", h.Login)
	g.POST("/auth/refresh", h.RefreshToken)
	g.POST("/auth/security-config", h.SecurityConfig)
	g.POST("/auth/captcha", h.Captcha)
	g.POST("/auth/email-code/send", h.SendEmailCode)
	g.POST("/auth/password-reset/code/send", h.SendPasswordResetCode)
	g.POST("/auth/password-reset/code/verify", h.VerifyPasswordResetCode)
	g.POST("/auth/password-reset/confirm", h.ResetPassword)
}

// RefreshToken POST /api/v1/auth/refresh 使用 Refresh Token 轮换双 Token。
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	req := new(api.RefreshTokenRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 401, "refresh token is required")
		return
	}
	result, err := h.biz.RefreshToken(req)
	if err != nil {
		response.Error(c, 401, "invalid or expired refresh token")
		return
	}
	response.OK(c, result)
}

// VerifyPasswordResetCode POST /api/v1/auth/password-reset/code/verify 验证密码重置二次验证码。
func (h *AuthHandler) VerifyPasswordResetCode(c *gin.Context) {
	req := new(api.PasswordResetVerifyRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请输入验证码")
		return
	}
	result, err := h.biz.VerifyPasswordResetCode(req)
	if err != nil {
		response.Error(c, 400, "验证码错误或已过期")
		return
	}
	response.OK(c, result)
}

// SendPasswordResetCode POST /api/v1/auth/password-reset/code/send 发送密码重置二次验证码。
func (h *AuthHandler) SendPasswordResetCode(c *gin.Context) {
	req := new(api.PasswordResetCodeRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请输入用户名")
		return
	}
	result, err := h.biz.SendPasswordResetCode(req)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidCredentials) {
			response.Error(c, 400, "用户不存在或已停用")
			return
		}
		if errors.Is(err, biz.ErrInvalidCaptcha) {
			response.Error(c, 400, "图形验证码错误或已过期")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// ResetPassword POST /api/v1/auth/password-reset/confirm 使用重置凭证修改密码。
func (h *AuthHandler) ResetPassword(c *gin.Context) {
	req := new(api.PasswordResetRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "重置密码参数不正确")
		return
	}
	if err := h.biz.ResetPassword(req); err != nil {
		if errors.Is(err, biz.ErrInvalidEmailCode) {
			response.Error(c, 400, "验证码错误或已过期")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// Login POST /api/v1/auth/login 用户登录。
func (h *AuthHandler) Login(c *gin.Context) {
	req := new(api.LoginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		h.biz.RecordInvalidLogin(req.Username, c.ClientIP(), "用户名和密码不能为空")
		response.Error(c, 400, "username and password are required")
		return
	}

	result, err := h.biz.Login(req, c.ClientIP())
	if err != nil {
		if errors.Is(err, biz.ErrInvalidEmailCode) {
			response.Error(c, 401, "邮件验证码错误或已过期")
			return
		}
		if errors.Is(err, biz.ErrInvalidCredentials) || errors.Is(err, biz.ErrInvalidCaptcha) {
			response.Error(c, 401, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, result)
}

// SecurityConfig POST /api/v1/auth/security-config 获取登录安全配置。
func (h *AuthHandler) SecurityConfig(c *gin.Context) {
	result, err := h.biz.SecurityConfig()
	if err != nil {
		response.Error(c, 500, "load login security config failed")
		return
	}
	response.OK(c, result)
}

// Captcha POST /api/v1/auth/captcha 生成图形验证码。
func (h *AuthHandler) Captcha(c *gin.Context) {
	result, err := h.biz.CreateCaptcha()
	if err != nil {
		response.Error(c, 500, "generate captcha failed")
		return
	}
	response.OK(c, result)
}

// SendEmailCode POST /api/v1/auth/email-code/send 按系统配置发送登录验证码。
func (h *AuthHandler) SendEmailCode(c *gin.Context) {
	req := new(api.EmailCodeRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "username and password are required")
		return
	}
	result, err := h.biz.SendEmailCode(req)
	if err != nil {
		if errors.Is(err, biz.ErrInvalidCredentials) || errors.Is(err, biz.ErrInvalidCaptcha) {
			response.Error(c, 400, err.Error())
			return
		}
		response.Error(c, 500, "send login verification code failed")
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
