package handler

import (
	"errors"
	"net"
	"net/url"
	"strings"

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
	g.POST("/auth/mfa/status", h.MFAStatus)
	g.POST("/auth/mfa/totp/setup", h.SetupTOTP)
	g.POST("/auth/mfa/totp/confirm", h.ConfirmTOTP)
	g.POST("/auth/mfa/totp/disable", h.DisableTOTP)
	g.POST("/auth/mfa/passkeys/register/begin", h.BeginPasskeyRegistration)
	g.POST("/auth/mfa/passkeys/register/finish", h.FinishPasskeyRegistration)
	g.POST("/auth/mfa/passkeys/delete", h.DeletePasskey)
}

// MFAStatus POST /api/v1/auth/mfa/status 查询当前用户 MFA 绑定状态。
func (h *AuthHandler) MFAStatus(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	result, err := h.biz.MFAStatus(userID)
	if err != nil {
		response.Error(c, 500, "查询 MFA 状态失败")
		return
	}
	response.OK(c, result)
}

// SetupTOTP POST /api/v1/auth/mfa/totp/setup 创建当前用户 TOTP 绑定信息。
func (h *AuthHandler) SetupTOTP(c *gin.Context) {
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	result, err := h.biz.BeginTOTPSetup(userID)
	if err != nil {
		response.Error(c, 500, "创建 TOTP 配置失败")
		return
	}
	response.OK(c, result)
}

// ConfirmTOTP POST /api/v1/auth/mfa/totp/confirm 验证并启用当前用户 TOTP。
func (h *AuthHandler) ConfirmTOTP(c *gin.Context) {
	req := new(api.TOTPConfirmRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请输入 TOTP 验证码")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	if err := h.biz.ConfirmTOTPSetup(userID, req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// DisableTOTP POST /api/v1/auth/mfa/totp/disable 验证并停用当前用户 TOTP。
func (h *AuthHandler) DisableTOTP(c *gin.Context) {
	req := new(api.TOTPConfirmRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请输入 TOTP 验证码")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	if err := h.biz.DisableTOTP(userID, req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
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
	g.POST("/auth/mfa/code/send", h.SendMFACode)
	g.POST("/auth/mfa/verify", h.VerifyMFA)
	g.POST("/auth/mfa/passkey/login/begin", h.BeginPasskeyLogin)
	g.POST("/auth/mfa/passkey/login/finish", h.FinishPasskeyLogin)
}

// SendMFACode POST /api/v1/auth/mfa/code/send 发送 MFA 登录验证码。
func (h *AuthHandler) SendMFACode(c *gin.Context) {
	req := new(api.MFACodeSendRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "MFA 登录会话不能为空")
		return
	}
	result, err := h.biz.SendMFACode(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// VerifyMFA POST /api/v1/auth/mfa/verify 使用验证码或 TOTP 完成 MFA 登录。
func (h *AuthHandler) VerifyMFA(c *gin.Context) {
	req := new(api.MFAVerifyRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "MFA 验证参数不完整")
		return
	}
	result, err := h.biz.VerifyMFA(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// BeginPasskeyLogin POST /api/v1/auth/mfa/passkey/login/begin 创建 Passkey 登录挑战。
func (h *AuthHandler) BeginPasskeyLogin(c *gin.Context) {
	req := new(api.PasskeyLoginBeginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "MFA 登录会话不能为空")
		return
	}
	result, err := h.biz.BeginPasskeyLogin(req, requestOrigin(c))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// FinishPasskeyLogin POST /api/v1/auth/mfa/passkey/login/finish 验证 Passkey 并完成 MFA 登录。
func (h *AuthHandler) FinishPasskeyLogin(c *gin.Context) {
	req := new(api.PasskeyLoginFinishRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "Passkey 登录响应不完整")
		return
	}
	result, err := h.biz.FinishPasskeyLogin(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// BeginPasskeyRegistration POST /api/v1/auth/mfa/passkeys/register/begin 创建 Passkey 注册挑战。
func (h *AuthHandler) BeginPasskeyRegistration(c *gin.Context) {
	req := new(api.PasskeyRegisterBeginRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "请输入 Passkey 名称")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	result, err := h.biz.BeginPasskeyRegistration(userID, req.Name, requestOrigin(c))
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// FinishPasskeyRegistration POST /api/v1/auth/mfa/passkeys/register/finish 验证并保存 Passkey。
func (h *AuthHandler) FinishPasskeyRegistration(c *gin.Context) {
	req := new(api.PasskeyRegisterFinishRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "Passkey 注册响应不完整")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	if err := h.biz.FinishPasskeyRegistration(userID, req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

// DeletePasskey POST /api/v1/auth/mfa/passkeys/delete 删除当前用户的 Passkey。
func (h *AuthHandler) DeletePasskey(c *gin.Context) {
	req := new(api.PasskeyDeleteRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "Passkey ID 不能为空")
		return
	}
	userID, ok := middleware.CurrentUserID(c)
	if !ok {
		response.Error(c, 401, "unauthorized")
		return
	}
	if err := h.biz.DeletePasskey(userID, req); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, nil)
}

func requestOrigin(c *gin.Context) string {
	if origin := c.GetHeader("Origin"); origin != "" {
		parsed, err := url.Parse(origin)
		requestHost := c.Request.Host
		if forwardedHost := strings.TrimSpace(strings.Split(c.GetHeader("X-Forwarded-Host"), ",")[0]); forwardedHost != "" {
			requestHost = forwardedHost
		}
		hostname := requestHost
		if host, _, splitErr := net.SplitHostPort(requestHost); splitErr == nil {
			hostname = host
		}
		if err == nil && strings.EqualFold(parsed.Hostname(), hostname) {
			return origin
		}
	}
	scheme := "http"
	if c.Request.TLS != nil || c.GetHeader("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}
	return scheme + "://" + c.Request.Host
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
