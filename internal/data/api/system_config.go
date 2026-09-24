package api

type SystemConfigRequest struct {
	SystemName                       string   `json:"system_name" binding:"required"`
	SystemDescription                string   `json:"system_description"`
	PasswordMinLength                int      `json:"password_min_length" binding:"min=1,max=128"`
	PasswordRequireDigit             bool     `json:"password_require_digit"`
	PasswordRequireLetter            bool     `json:"password_require_letter"`
	PasswordRequireUpper             bool     `json:"password_require_upper"`
	PasswordRequireLower             bool     `json:"password_require_lower"`
	PasswordRequireSpecial           bool     `json:"password_require_special"`
	SessionTimeoutMinutes            int      `json:"session_timeout_minutes" binding:"min=1,max=10080"`
	CaptchaEnabled                   bool     `json:"captcha_enabled"`
	CaptchaExpireMinutes             int      `json:"captcha_expire_minutes" binding:"min=1,max=60"`
	LoginVerificationEnabled         bool     `json:"login_verification_enabled"`
	LoginVerificationMethods         []string `json:"login_verification_methods" binding:"omitempty,dive,oneof=email webhook passkey totp"`
	MaxLoginFailures                 int      `json:"max_login_failures" binding:"min=1,max=100"`
	AccountLockMinutes               int      `json:"account_lock_minutes" binding:"min=1,max=10080"`
	VerificationEmailEnabled         bool     `json:"verification_email_enabled"`
	VerificationEmailExpireMinutes   int      `json:"verification_email_expire_minutes" binding:"min=1,max=60"`
	VerificationEmailSenderName      string   `json:"verification_email_sender_name"`
	VerificationSMTPHost             string   `json:"verification_smtp_host"`
	VerificationSMTPPort             int      `json:"verification_smtp_port" binding:"min=1,max=65535"`
	VerificationSMTPUsername         string   `json:"verification_smtp_username"`
	VerificationSMTPPassword         string   `json:"verification_smtp_password"`
	VerificationSMTPTLS              bool     `json:"verification_smtp_tls"`
	VerificationWebhookEnabled       bool     `json:"verification_webhook_enabled"`
	VerificationWebhookURL           string   `json:"verification_webhook_url"`
	VerificationWebhookSecret        string   `json:"verification_webhook_secret"`
	VerificationPasskeyEnabled       bool     `json:"verification_passkey_enabled"`
	VerificationTOTPEnabled          bool     `json:"verification_totp_enabled"`
	DirectVerificationEmailEnabled   bool     `json:"direct_verification_email_enabled"`
	DirectVerificationPasskeyEnabled bool     `json:"direct_verification_passkey_enabled"`
	DirectVerificationTOTPEnabled    bool     `json:"direct_verification_totp_enabled"`
}

type SystemConfigResponse struct {
	SystemName                       string   `json:"system_name"`
	SystemDescription                string   `json:"system_description"`
	PasswordMinLength                int      `json:"password_min_length"`
	PasswordRequireDigit             bool     `json:"password_require_digit"`
	PasswordRequireLetter            bool     `json:"password_require_letter"`
	PasswordRequireUpper             bool     `json:"password_require_upper"`
	PasswordRequireLower             bool     `json:"password_require_lower"`
	PasswordRequireSpecial           bool     `json:"password_require_special"`
	SessionTimeoutMinutes            int      `json:"session_timeout_minutes"`
	CaptchaEnabled                   bool     `json:"captcha_enabled"`
	CaptchaExpireMinutes             int      `json:"captcha_expire_minutes"`
	LoginVerificationEnabled         bool     `json:"login_verification_enabled"`
	LoginVerificationMethods         []string `json:"login_verification_methods"`
	LoginVerificationMethod          string   `json:"-"`
	MaxLoginFailures                 int      `json:"max_login_failures"`
	AccountLockMinutes               int      `json:"account_lock_minutes"`
	VerificationEmailEnabled         bool     `json:"verification_email_enabled"`
	VerificationEmailExpireMinutes   int      `json:"verification_email_expire_minutes"`
	VerificationEmailSenderName      string   `json:"verification_email_sender_name"`
	VerificationSMTPHost             string   `json:"verification_smtp_host"`
	VerificationSMTPPort             int      `json:"verification_smtp_port"`
	VerificationSMTPUsername         string   `json:"verification_smtp_username"`
	VerificationSMTPPassword         string   `json:"verification_smtp_password"`
	VerificationSMTPTLS              bool     `json:"verification_smtp_tls"`
	VerificationWebhookEnabled       bool     `json:"verification_webhook_enabled"`
	VerificationWebhookURL           string   `json:"verification_webhook_url"`
	VerificationWebhookSecret        string   `json:"verification_webhook_secret"`
	VerificationPasskeyEnabled       bool     `json:"verification_passkey_enabled"`
	VerificationTOTPEnabled          bool     `json:"verification_totp_enabled"`
	DirectVerificationEmailEnabled   bool     `json:"direct_verification_email_enabled"`
	DirectVerificationPasskeyEnabled bool     `json:"direct_verification_passkey_enabled"`
	DirectVerificationTOTPEnabled    bool     `json:"direct_verification_totp_enabled"`
}

type SystemConfigEmailTestRequest struct {
	Recipient string `json:"recipient" binding:"required,email"`
}
