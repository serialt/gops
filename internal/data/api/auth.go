package api

type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	EmailCode   string `json:"email_code"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type ProfileUpdateRequest struct {
	Username        string `json:"username" binding:"required"`
	Nickname        string `json:"nickname" binding:"required"`
	Email           string `json:"email" binding:"omitempty,email"`
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type LoginSecurityConfigResponse struct {
	CaptchaEnabled           bool   `json:"captcha_enabled"`
	LoginVerificationEnabled bool   `json:"login_verification_enabled"`
	LoginVerificationMethod  string `json:"login_verification_method"`
}

type CaptchaResponse struct {
	CaptchaID string `json:"captcha_id"`
	Image     string `json:"image"`
}

type EmailCodeRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type EmailCodeResponse struct {
	Required          bool `json:"required"`
	Sent              bool `json:"sent"`
	RetryAfterSeconds int  `json:"retry_after_seconds"`
	ExpiresInSeconds  int  `json:"expires_in_seconds"`
}
