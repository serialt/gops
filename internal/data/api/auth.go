package api

import "encoding/json"

type LoginRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
	EmailCode   string `json:"email_code"`
}

type LoginResponse struct {
	Token            string   `json:"token"`
	AccessToken      string   `json:"access_token"`
	RefreshToken     string   `json:"refresh_token"`
	ExpiresIn        int64    `json:"expires_in"`
	RefreshExpiresIn int64    `json:"refresh_expires_in"`
	MFARequired      bool     `json:"mfa_required"`
	MFAToken         string   `json:"mfa_token"`
	MFAMethods       []string `json:"mfa_methods"`
}

type MFAVerifyRequest struct {
	MFAToken string `json:"mfa_token" binding:"required"`
	Method   string `json:"method" binding:"required"`
	Code     string `json:"code"`
}

type MFACodeSendRequest struct {
	MFAToken string `json:"mfa_token" binding:"required"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
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

type PasswordResetCodeRequest struct {
	Username    string `json:"username" binding:"required"`
	CaptchaID   string `json:"captcha_id"`
	CaptchaCode string `json:"captcha_code"`
}

type PasswordResetVerifyRequest struct {
	Username string `json:"username" binding:"required"`
	Code     string `json:"code" binding:"required"`
}

type PasswordResetVerifyResponse struct {
	ResetToken string `json:"reset_token"`
}

type PasswordResetRequest struct {
	ResetToken  string `json:"reset_token" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

type MFAStatusResponse struct {
	CodeAvailable bool                  `json:"code_available"`
	TOTPEnabled   bool                  `json:"totp_enabled"`
	Passkeys      []*MFAPasskeyResponse `json:"passkeys"`
}

type MFAPasskeyResponse struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type TOTPSetupResponse struct {
	Secret string `json:"secret"`
	URI    string `json:"uri"`
}

type TOTPConfirmRequest struct {
	Code string `json:"code" binding:"required"`
}

type PasskeyRegisterBeginRequest struct {
	Name string `json:"name" binding:"required"`
}

type PasskeyRegisterBeginResponse struct {
	RegistrationToken string `json:"registration_token"`
	Options           any    `json:"options"`
}

type PasskeyRegisterFinishRequest struct {
	RegistrationToken string          `json:"registration_token" binding:"required"`
	Name              string          `json:"name" binding:"required"`
	Credential        json.RawMessage `json:"credential" binding:"required"`
}

type PasskeyDeleteRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type PasskeyLoginBeginRequest struct {
	MFAToken string `json:"mfa_token" binding:"required"`
}

type PasskeyLoginBeginResponse struct {
	Options any `json:"options"`
}

type PasskeyLoginFinishRequest struct {
	MFAToken   string          `json:"mfa_token" binding:"required"`
	Credential json.RawMessage `json:"credential" binding:"required"`
}
