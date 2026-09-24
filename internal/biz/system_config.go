package biz

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"net"
	"net/http"
	"net/mail"
	"net/smtp"
	"strconv"
	"strings"
	"time"

	"gops/internal/config"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var systemConfigKeys = []string{
	"system_name", "system_description", "password_min_length", "password_require_digit",
	"password_require_letter", "password_require_upper", "password_require_lower", "password_require_special",
	"session_timeout_minutes", "captcha_enabled", "max_login_failures", "account_lock_minutes",
	"captcha_expire_minutes", "login_verification_enabled", "login_verification_method",
	"verification_email_enabled", "verification_email_expire_minutes", "verification_email_sender_name", "verification_smtp_host", "verification_smtp_port",
	"verification_smtp_username", "verification_smtp_password", "verification_smtp_tls", "verification_webhook_enabled",
	"verification_webhook_url", "verification_webhook_secret",
	"verification_passkey_enabled", "verification_totp_enabled",
	"direct_verification_email_enabled", "direct_verification_passkey_enabled", "direct_verification_totp_enabled",
}

var ErrInvalidVerificationConfig = errors.New("selected login verification method is not enabled")

type SystemConfigBiz struct {
	db     *gorm.DB
	aesKey string
}

func NewSystemConfigBiz(db *gorm.DB, cfg *config.Config) *SystemConfigBiz {
	return &SystemConfigBiz{db: db, aesKey: cfg.Server.AESKey}
}

func (b *SystemConfigBiz) Get() (*api.SystemConfigResponse, error) {
	var rows []model.SystemConfig
	if err := b.db.Where("config_key IN ?", systemConfigKeys).Find(&rows).Error; err != nil {
		return nil, err
	}
	values := map[string]string{
		"system_name": "Gops", "system_description": "", "password_min_length": "8",
		"password_require_digit": "false", "password_require_letter": "false", "password_require_upper": "false",
		"password_require_lower": "false", "password_require_special": "false",
		"session_timeout_minutes": "30", "captcha_enabled": "false", "captcha_expire_minutes": "1",
		"login_verification_enabled": "false", "login_verification_method": "",
		"max_login_failures": "5", "account_lock_minutes": "30",
		"verification_email_enabled": "true", "verification_email_expire_minutes": "5", "verification_email_sender_name": "", "verification_smtp_host": "",
		"verification_smtp_port": "465", "verification_smtp_username": "", "verification_smtp_password": "", "verification_smtp_tls": "true",
		"verification_webhook_enabled": "false", "verification_webhook_url": "", "verification_webhook_secret": "",
		"verification_passkey_enabled": "true", "verification_totp_enabled": "true",
		"direct_verification_email_enabled":   "true",
		"direct_verification_passkey_enabled": "true", "direct_verification_totp_enabled": "true",
	}
	for _, row := range rows {
		values[row.ConfigKey] = row.ConfigValue
	}
	if !parseConfigBool(values["verification_email_enabled"]) && !parseConfigBool(values["verification_webhook_enabled"]) && !parseConfigBool(values["verification_passkey_enabled"]) && !parseConfigBool(values["verification_totp_enabled"]) {
		values["verification_email_enabled"] = "true"
	}
	for _, key := range []string{"verification_smtp_password", "verification_webhook_secret"} {
		decrypted, err := cryptox.Decrypt(b.aesKey, values[key])
		if err != nil {
			return nil, err
		}
		values[key] = decrypted
	}
	return &api.SystemConfigResponse{
		SystemName: values["system_name"], SystemDescription: values["system_description"],
		PasswordMinLength:                parseConfigInt(values["password_min_length"], 8),
		PasswordRequireDigit:             parseConfigBool(values["password_require_digit"]),
		PasswordRequireLetter:            parseConfigBool(values["password_require_letter"]),
		PasswordRequireUpper:             parseConfigBool(values["password_require_upper"]),
		PasswordRequireLower:             parseConfigBool(values["password_require_lower"]),
		PasswordRequireSpecial:           parseConfigBool(values["password_require_special"]),
		SessionTimeoutMinutes:            parseConfigInt(values["session_timeout_minutes"], 30),
		CaptchaEnabled:                   parseConfigBool(values["captcha_enabled"]),
		CaptchaExpireMinutes:             parseConfigInt(values["captcha_expire_minutes"], 1),
		LoginVerificationEnabled:         parseConfigBool(values["login_verification_enabled"]),
		LoginVerificationMethod:          values["login_verification_method"],
		MaxLoginFailures:                 parseConfigInt(values["max_login_failures"], 5),
		AccountLockMinutes:               parseConfigInt(values["account_lock_minutes"], 30),
		VerificationEmailEnabled:         parseConfigBool(values["verification_email_enabled"]),
		VerificationEmailExpireMinutes:   parseConfigInt(values["verification_email_expire_minutes"], 5),
		VerificationEmailSenderName:      values["verification_email_sender_name"],
		VerificationSMTPHost:             values["verification_smtp_host"],
		VerificationSMTPPort:             parseConfigInt(values["verification_smtp_port"], 465),
		VerificationSMTPUsername:         values["verification_smtp_username"],
		VerificationSMTPPassword:         values["verification_smtp_password"],
		VerificationSMTPTLS:              parseConfigBool(values["verification_smtp_tls"]),
		VerificationWebhookEnabled:       parseConfigBool(values["verification_webhook_enabled"]),
		VerificationWebhookURL:           values["verification_webhook_url"],
		VerificationWebhookSecret:        values["verification_webhook_secret"],
		VerificationPasskeyEnabled:       parseConfigBool(values["verification_passkey_enabled"]),
		VerificationTOTPEnabled:          parseConfigBool(values["verification_totp_enabled"]),
		DirectVerificationEmailEnabled:   parseConfigBool(values["direct_verification_email_enabled"]),
		DirectVerificationPasskeyEnabled: parseConfigBool(values["direct_verification_passkey_enabled"]),
		DirectVerificationTOTPEnabled:    parseConfigBool(values["direct_verification_totp_enabled"]),
	}, nil
}

func (b *SystemConfigBiz) Update(req *api.SystemConfigRequest) (*api.SystemConfigResponse, error) {
	if !req.VerificationEmailEnabled && !req.VerificationWebhookEnabled && !req.VerificationPasskeyEnabled && !req.VerificationTOTPEnabled {
		return nil, ErrInvalidVerificationConfig
	}
	if req.LoginVerificationEnabled {
		if req.LoginVerificationMethod == "" && (req.VerificationEmailEnabled || req.VerificationWebhookEnabled) {
			return nil, ErrInvalidVerificationConfig
		}
		if req.LoginVerificationMethod == "email" && !req.VerificationEmailEnabled {
			return nil, ErrInvalidVerificationConfig
		}
		if req.LoginVerificationMethod == "webhook" && (!req.VerificationWebhookEnabled || req.VerificationWebhookURL == "") {
			return nil, ErrInvalidVerificationConfig
		}
	}
	if !req.VerificationPasskeyEnabled {
		req.DirectVerificationPasskeyEnabled = false
	}
	if !req.VerificationEmailEnabled {
		req.DirectVerificationEmailEnabled = false
	}
	if !req.VerificationTOTPEnabled {
		req.DirectVerificationTOTPEnabled = false
	}
	smtpPassword, err := cryptox.Encrypt(b.aesKey, req.VerificationSMTPPassword)
	if err != nil {
		return nil, err
	}
	webhookSecret, err := cryptox.Encrypt(b.aesKey, req.VerificationWebhookSecret)
	if err != nil {
		return nil, err
	}
	configs := []model.SystemConfig{
		{ConfigKey: "system_name", ConfigValue: req.SystemName},
		{ConfigKey: "system_description", ConfigValue: req.SystemDescription},
		{ConfigKey: "password_min_length", ConfigValue: strconv.Itoa(req.PasswordMinLength)},
		{ConfigKey: "password_require_digit", ConfigValue: strconv.FormatBool(req.PasswordRequireDigit)},
		{ConfigKey: "password_require_letter", ConfigValue: strconv.FormatBool(req.PasswordRequireLetter)},
		{ConfigKey: "password_require_upper", ConfigValue: strconv.FormatBool(req.PasswordRequireUpper)},
		{ConfigKey: "password_require_lower", ConfigValue: strconv.FormatBool(req.PasswordRequireLower)},
		{ConfigKey: "password_require_special", ConfigValue: strconv.FormatBool(req.PasswordRequireSpecial)},
		{ConfigKey: "session_timeout_minutes", ConfigValue: strconv.Itoa(req.SessionTimeoutMinutes)},
		{ConfigKey: "captcha_enabled", ConfigValue: strconv.FormatBool(req.CaptchaEnabled)},
		{ConfigKey: "captcha_expire_minutes", ConfigValue: strconv.Itoa(req.CaptchaExpireMinutes)},
		{ConfigKey: "login_verification_enabled", ConfigValue: strconv.FormatBool(req.LoginVerificationEnabled)},
		{ConfigKey: "login_verification_method", ConfigValue: req.LoginVerificationMethod},
		{ConfigKey: "max_login_failures", ConfigValue: strconv.Itoa(req.MaxLoginFailures)},
		{ConfigKey: "account_lock_minutes", ConfigValue: strconv.Itoa(req.AccountLockMinutes)},
		{ConfigKey: "verification_email_enabled", ConfigValue: strconv.FormatBool(req.VerificationEmailEnabled)},
		{ConfigKey: "verification_email_expire_minutes", ConfigValue: strconv.Itoa(req.VerificationEmailExpireMinutes)},
		{ConfigKey: "verification_email_sender_name", ConfigValue: req.VerificationEmailSenderName},
		{ConfigKey: "verification_smtp_host", ConfigValue: req.VerificationSMTPHost},
		{ConfigKey: "verification_smtp_port", ConfigValue: strconv.Itoa(req.VerificationSMTPPort)},
		{ConfigKey: "verification_smtp_username", ConfigValue: req.VerificationSMTPUsername},
		{ConfigKey: "verification_smtp_password", ConfigValue: smtpPassword},
		{ConfigKey: "verification_smtp_tls", ConfigValue: strconv.FormatBool(req.VerificationSMTPTLS)},
		{ConfigKey: "verification_webhook_enabled", ConfigValue: strconv.FormatBool(req.VerificationWebhookEnabled)},
		{ConfigKey: "verification_webhook_url", ConfigValue: req.VerificationWebhookURL},
		{ConfigKey: "verification_webhook_secret", ConfigValue: webhookSecret},
		{ConfigKey: "verification_passkey_enabled", ConfigValue: strconv.FormatBool(req.VerificationPasskeyEnabled)},
		{ConfigKey: "verification_totp_enabled", ConfigValue: strconv.FormatBool(req.VerificationTOTPEnabled)},
		{ConfigKey: "direct_verification_email_enabled", ConfigValue: strconv.FormatBool(req.DirectVerificationEmailEnabled)},
		{ConfigKey: "direct_verification_passkey_enabled", ConfigValue: strconv.FormatBool(req.DirectVerificationPasskeyEnabled)},
		{ConfigKey: "direct_verification_totp_enabled", ConfigValue: strconv.FormatBool(req.DirectVerificationTOTPEnabled)},
	}
	if err := b.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "config_key"}},
		DoUpdates: clause.AssignmentColumns([]string{"config_value", "updated_at"}),
	}).Create(&configs).Error; err != nil {
		return nil, err
	}
	return b.Get()
}

func (b *SystemConfigBiz) TestEmail(req *api.SystemConfigEmailTestRequest) error {
	return b.sendEmail(req.Recipient, "SMTP 配置测试", "SMTP 配置测试邮件发送成功。", false)
}

func (b *SystemConfigBiz) SendLoginVerification(config *api.SystemConfigResponse, username, email, code string) error {
	return b.sendVerification(config, username, email, code, "login_verification", "登录验证码")
}

func (b *SystemConfigBiz) SendEmailLoginCode(config *api.SystemConfigResponse, email, code string) error {
	if !config.VerificationEmailEnabled || strings.TrimSpace(email) == "" {
		return ErrInvalidVerificationConfig
	}
	content, err := renderVerificationEmail("登录验证码", code, config.VerificationEmailExpireMinutes)
	if err != nil {
		return err
	}
	return b.sendEmail(email, "登录验证码", content, true)
}

func (b *SystemConfigBiz) SendPasswordResetVerification(config *api.SystemConfigResponse, username, email, code string) error {
	return b.sendVerification(config, username, email, code, "password_reset", "密码重置验证码")
}

func (b *SystemConfigBiz) sendVerification(config *api.SystemConfigResponse, username, email, code, event, subject string) error {
	if !config.LoginVerificationEnabled {
		return nil
	}
	if config.LoginVerificationMethod == "webhook" {
		if !config.VerificationWebhookEnabled || config.VerificationWebhookURL == "" {
			return ErrInvalidVerificationConfig
		}
		payload, err := json.Marshal(map[string]any{
			"event": event, "username": username, "email": email,
			"code": code, "expires_in": 300,
		})
		if err != nil {
			return err
		}
		req, err := http.NewRequest(http.MethodPost, config.VerificationWebhookURL, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
		if config.VerificationWebhookSecret != "" {
			mac := hmac.New(sha256.New, []byte(config.VerificationWebhookSecret))
			_, _ = mac.Write(payload)
			req.Header.Set("X-Gops-Signature", hex.EncodeToString(mac.Sum(nil)))
		}
		response, err := (&http.Client{Timeout: 10 * time.Second}).Do(req)
		if err != nil {
			return err
		}
		defer response.Body.Close()
		if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
			return fmt.Errorf("webhook returned status %d", response.StatusCode)
		}
		return nil
	}
	if config.LoginVerificationMethod != "email" || !config.VerificationEmailEnabled || email == "" {
		return ErrInvalidVerificationConfig
	}
	content, err := renderVerificationEmail(subject, code, config.VerificationEmailExpireMinutes)
	if err != nil {
		return err
	}
	return b.sendEmail(email, subject, content, true)
}

func (b *SystemConfigBiz) sendEmail(recipient, subject, content string, htmlContent bool) error {
	config, err := b.Get()
	if err != nil {
		return err
	}
	if config.VerificationSMTPHost == "" || config.VerificationSMTPUsername == "" {
		return errors.New("SMTP host and username are required")
	}
	address := net.JoinHostPort(config.VerificationSMTPHost, strconv.Itoa(config.VerificationSMTPPort))
	client, err := newSMTPClient(address, config.VerificationSMTPHost, config.VerificationSMTPTLS, config.VerificationSMTPPort)
	if err != nil {
		return err
	}
	defer client.Quit()
	if config.VerificationSMTPUsername != "" {
		auth := smtp.PlainAuth("", config.VerificationSMTPUsername, config.VerificationSMTPPassword, config.VerificationSMTPHost)
		if err := client.Auth(auth); err != nil {
			return err
		}
	}
	if err := client.Mail(config.VerificationSMTPUsername); err != nil {
		return err
	}
	if err := client.Rcpt(recipient); err != nil {
		return err
	}
	body, err := client.Data()
	if err != nil {
		return err
	}
	recipientHeader := (&mail.Address{Address: cleanEmailHeader(recipient)}).String()
	senderHeader := (&mail.Address{Name: cleanEmailHeader(config.VerificationEmailSenderName), Address: cleanEmailHeader(config.VerificationSMTPUsername)}).String()
	subjectHeader := mime.QEncoding.Encode("UTF-8", cleanEmailHeader(subject))
	messageBody := strings.ReplaceAll(strings.ReplaceAll(content, "\r\n", "\n"), "\n", "\r\n")
	contentType := "text/plain"
	if htmlContent {
		contentType = "text/html"
	}
	_, writeErr := fmt.Fprintf(body, "To: %s\r\nFrom: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: %s; charset=UTF-8\r\nContent-Transfer-Encoding: 8bit\r\n\r\n%s\r\n", recipientHeader, senderHeader, subjectHeader, contentType, messageBody)
	closeErr := body.Close()
	if writeErr != nil {
		return writeErr
	}

	return closeErr
}

func cleanEmailHeader(value string) string {
	return strings.TrimSpace(strings.NewReplacer("\r", "", "\n", "").Replace(value))
}

func newSMTPClient(address, host string, tlsEnabled bool, port int) (*smtp.Client, error) {
	if tlsEnabled && port == 465 {
		connection, err := tls.Dial("tcp", address, &tls.Config{ServerName: host, MinVersion: tls.VersionTLS12})
		if err != nil {
			return nil, err
		}
		return smtp.NewClient(connection, host)
	}
	client, err := smtp.Dial(address)
	if err != nil {
		return nil, err
	}
	if tlsEnabled {
		if ok, _ := client.Extension("STARTTLS"); !ok {
			client.Close()
			return nil, errors.New("SMTP server does not support STARTTLS")
		}
		if err := client.StartTLS(&tls.Config{ServerName: host, MinVersion: tls.VersionTLS12}); err != nil {
			client.Close()
			return nil, err
		}
	}
	return client, nil
}

func parseConfigInt(value string, fallback int) int {
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 {
		return fallback
	}
	return parsed
}

func parseConfigBool(value string) bool {
	parsed, _ := strconv.ParseBool(value)
	return parsed
}
