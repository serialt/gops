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
}

func (h *AuthHandler) RegisterRoutesNologin(g *gin.RouterGroup) {
	g.POST("/auth/login", h.Login)
	g.POST("/auth/security-config", h.SecurityConfig)
	g.POST("/auth/captcha", h.Captcha)
	g.POST("/auth/email-code/send", h.SendEmailCode)
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
