package biz

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"html"
	"log/slog"
	"math/big"
	"strings"
	"sync"
	"time"
	"unicode"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"
	"gops/internal/pkg/cryptox"

	qrcode "github.com/skip2/go-qrcode"
	"gorm.io/gorm"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrGenerateToken      = errors.New("generate token failed")
	ErrInvalidCaptcha     = errors.New("invalid or expired captcha")
	ErrInvalidEmailCode   = errors.New("invalid or expired email code")
)

type verificationValue struct {
	code               string
	expiresAt          time.Time
	challengeExpiresAt time.Time
	attempts           int
}

type loginMFAChallenge struct {
	userID    uint
	username  string
	ip        string
	methods   []string
	expiresAt time.Time
}

type AuthBiz struct {
	db                   *gorm.DB
	jwt                  *auth.JWTManager
	casbin               *casbinx.Service
	config               *SystemConfigBiz
	mu                   sync.Mutex
	captchas             map[string]verificationValue
	emailCodes           map[string]verificationValue
	resetCodes           map[string]verificationValue
	resetTokens          map[string]verificationValue
	pendingTOTP          map[uint]verificationValue
	mfaChallenges        map[string]loginMFAChallenge
	passkeyRegistrations map[string]passkeySession
	passkeyLogins        map[string]passkeySession
	directPasskeyLogins  map[string]passkeySession
	directEmailLogins    map[string]loginMFAChallenge
}

func NewAuthBiz(db *gorm.DB, jwt *auth.JWTManager, casbinService *casbinx.Service, config *SystemConfigBiz) *AuthBiz {
	return &AuthBiz{db: db, jwt: jwt, casbin: casbinService, config: config, captchas: make(map[string]verificationValue), emailCodes: make(map[string]verificationValue), resetCodes: make(map[string]verificationValue), resetTokens: make(map[string]verificationValue), pendingTOTP: make(map[uint]verificationValue), mfaChallenges: make(map[string]loginMFAChallenge), passkeyRegistrations: make(map[string]passkeySession), passkeyLogins: make(map[string]passkeySession), directPasskeyLogins: make(map[string]passkeySession), directEmailLogins: make(map[string]loginMFAChallenge)}
}

func (b *AuthBiz) Login(req *api.LoginRequest, ip string) (*api.LoginResponse, error) {
	user, err := b.authenticate(req.Username, req.Password)
	if err != nil {
		b.recordLogin(0, req.Username, ip, "failed", "用户名或密码错误")
		return nil, ErrInvalidCredentials
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	if config.CaptchaEnabled && !b.verifyCaptcha(req.CaptchaID, req.CaptchaCode, true) {
		b.recordLogin(user.ID, user.Username, ip, "failed", "图形验证码错误或已过期")
		return nil, ErrInvalidCaptcha
	}
	methods, err := b.availableMFAMethods(uint(user.ID), config, user)
	if err != nil {
		return nil, err
	}
	if len(methods) > 0 {
		token, tokenErr := randomToken()
		if tokenErr != nil {
			return nil, tokenErr
		}
		b.mu.Lock()
		b.mfaChallenges[token] = loginMFAChallenge{userID: uint(user.ID), username: user.Username, ip: ip, methods: methods, expiresAt: time.Now().Add(5 * time.Minute)}
		b.mu.Unlock()
		return &api.LoginResponse{MFARequired: true, MFAToken: token, MFAMethods: methods}, nil
	}

	tokens, err := b.jwt.GeneratePair(uint(user.ID))
	if err != nil {
		b.recordLogin(user.ID, user.Username, ip, "failed", "生成登录令牌失败")
		return nil, ErrGenerateToken
	}
	b.recordLogin(user.ID, user.Username, ip, "success", "登录成功")
	return tokenResponse(tokens), nil
}

func (b *AuthBiz) availableMFAMethods(userID uint, config *api.SystemConfigResponse, user *model.User) ([]string, error) {
	methods := make([]string, 0, 3)
	var emailMFA model.UserMfaEmail
	emailEnabled := config.VerificationEmailEnabled && strings.TrimSpace(user.Email) != "" && b.db.Where("user_id = ? AND mfa_enabled = ?", userID, 1).First(&emailMFA).Error == nil
	webhookEnabled := config.LoginVerificationEnabled && config.LoginVerificationMethod == "webhook" && config.VerificationWebhookEnabled
	if emailEnabled || webhookEnabled {
		methods = append(methods, "code")
	}
	var count int64
	if err := b.db.Model(&model.UserMfaTotp{}).Where("user_id = ? AND enabled = ?", userID, 1).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		methods = append(methods, "totp")
	}
	if err := b.db.Model(&model.UserMfaPasskey{}).Where("user_id = ? AND enabled = ?", userID, 1).Count(&count).Error; err != nil {
		return nil, err
	}
	if count > 0 {
		methods = append(methods, "passkey")
	}
	return methods, nil
}

func (b *AuthBiz) RefreshToken(req *api.RefreshTokenRequest) (*api.LoginResponse, error) {
	claims, err := b.jwt.ParseRefresh(req.RefreshToken)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	var user model.User
	if err := b.db.Select("id", "status").First(&user, claims.UserID).Error; err != nil || !user.Status {
		return nil, ErrInvalidCredentials
	}
	tokens, err := b.jwt.GeneratePair(claims.UserID)
	if err != nil {
		return nil, ErrGenerateToken
	}
	return tokenResponse(tokens), nil
}

func tokenResponse(tokens *auth.TokenPair) *api.LoginResponse {
	return &api.LoginResponse{Token: tokens.AccessToken, AccessToken: tokens.AccessToken, RefreshToken: tokens.RefreshToken, ExpiresIn: tokens.ExpiresIn, RefreshExpiresIn: tokens.RefreshExpiresIn}
}

func (b *AuthBiz) SecurityConfig() (*api.LoginSecurityConfigResponse, error) {
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	return &api.LoginSecurityConfigResponse{
		CaptchaEnabled: config.CaptchaEnabled, LoginVerificationEnabled: config.LoginVerificationEnabled,
		LoginVerificationMethod: config.LoginVerificationMethod,
	}, nil
}

func (b *AuthBiz) CreateCaptcha() (*api.CaptchaResponse, error) {
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	id, err := randomToken()
	if err != nil {
		return nil, err
	}
	code, err := randomDigits(4)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.captchas[id] = verificationValue{code: code, expiresAt: time.Now().Add(time.Duration(config.CaptchaExpireMinutes) * time.Minute)}
	b.mu.Unlock()
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="120" height="40" viewBox="0 0 120 40"><rect width="120" height="40" rx="7" fill="#f3f5ff"/><path d="M8 30L32 8M48 35L75 5M83 34L112 10" stroke="#c7d2f7" stroke-width="2"/><text x="60" y="27" text-anchor="middle" font-family="monospace" font-size="22" font-weight="700" letter-spacing="6" fill="#536fe8">%s</text></svg>`, html.EscapeString(code))
	return &api.CaptchaResponse{CaptchaID: id, Image: "data:image/svg+xml;base64," + base64.StdEncoding.EncodeToString([]byte(svg))}, nil
}

func (b *AuthBiz) SendEmailCode(req *api.EmailCodeRequest) (*api.EmailCodeResponse, error) {
	user, err := b.authenticate(req.Username, req.Password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	verificationRequired := b.loginVerificationRequired(config, user)
	if config.CaptchaEnabled && !b.hasActiveEmailCode(user.Username) && !b.verifyCaptcha(req.CaptchaID, req.CaptchaCode, verificationRequired) {
		return nil, ErrInvalidCaptcha
	}
	if !verificationRequired {
		return &api.EmailCodeResponse{Required: false, Sent: false}, nil
	}
	code, err := randomDigits(6)
	if err != nil {
		return nil, err
	}
	expiresIn := 5 * time.Minute
	if config.LoginVerificationMethod == "email" {
		expiresIn = time.Duration(config.VerificationEmailExpireMinutes) * time.Minute
	}
	if retryAfter, expiresAfter := b.reserveEmailCode(user.Username, code, expiresIn); retryAfter > 0 {
		return &api.EmailCodeResponse{Required: true, Sent: false, RetryAfterSeconds: retryAfter, ExpiresInSeconds: expiresAfter}, nil
	}
	go func(username, email, verificationCode string) {
		if err := b.config.SendLoginVerification(config, username, email, verificationCode); err != nil {
			b.discardEmailCode(username, verificationCode)
			slog.Error("发送登录验证码失败", "error", err, "username", username, "method", config.LoginVerificationMethod)
		}
	}(user.Username, user.Email, code)
	return &api.EmailCodeResponse{Required: true, Sent: true, RetryAfterSeconds: int(expiresIn.Seconds()), ExpiresInSeconds: int(expiresIn.Seconds())}, nil
}

func (b *AuthBiz) SendPasswordResetCode(req *api.PasswordResetCodeRequest) (*api.EmailCodeResponse, error) {
	var user model.User
	if err := b.db.Where("username = ? AND status = ?", strings.TrimSpace(req.Username), true).First(&user).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	resetMethod := passwordResetVerificationMethod(config, &user)
	if resetMethod == "" {
		return nil, errors.New("验证码配置中没有适用于当前用户的发送方式")
	}
	resetConfig := *config
	resetConfig.LoginVerificationEnabled = true
	resetConfig.LoginVerificationMethod = resetMethod
	if config.CaptchaEnabled && !b.verifyCaptcha(req.CaptchaID, req.CaptchaCode, true) {
		return nil, ErrInvalidCaptcha
	}
	expiresIn := 5 * time.Minute
	if resetMethod == "email" {
		expiresIn = time.Duration(config.VerificationEmailExpireMinutes) * time.Minute
	}
	code, err := randomDigits(6)
	if err != nil {
		return nil, err
	}
	if retryAfter, expiresAfter := b.reserveResetCode(user.Username, code, expiresIn); retryAfter > 0 {
		return &api.EmailCodeResponse{Required: true, Sent: false, RetryAfterSeconds: retryAfter, ExpiresInSeconds: expiresAfter}, nil
	}
	go func(username, email, verificationCode string) {
		if err := b.config.SendPasswordResetVerification(&resetConfig, username, email, verificationCode); err != nil {
			b.discardResetCode(username, verificationCode)
			slog.Error("发送密码重置验证码失败", "error", err, "username", username, "method", resetMethod)
		}
	}(user.Username, user.Email, code)
	return &api.EmailCodeResponse{Required: true, Sent: true, RetryAfterSeconds: int(expiresIn.Seconds()), ExpiresInSeconds: int(expiresIn.Seconds())}, nil
}

func passwordResetVerificationMethod(config *api.SystemConfigResponse, user *model.User) string {
	if config.VerificationEmailEnabled && strings.TrimSpace(user.Email) != "" {
		return "email"
	}
	if config.VerificationWebhookEnabled {
		return "webhook"
	}
	return ""
}

func (b *AuthBiz) VerifyPasswordResetCode(req *api.PasswordResetVerifyRequest) (*api.PasswordResetVerifyResponse, error) {
	username := strings.TrimSpace(req.Username)
	if !b.verifyResetCode(username, req.Code) {
		return nil, ErrInvalidEmailCode
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.resetTokens[token] = verificationValue{code: username, expiresAt: time.Now().Add(10 * time.Minute)}
	b.mu.Unlock()
	return &api.PasswordResetVerifyResponse{ResetToken: token}, nil
}

func (b *AuthBiz) ResetPassword(req *api.PasswordResetRequest) error {
	config, err := b.config.Get()
	if err != nil {
		return err
	}
	if err := validateProfilePassword(req.NewPassword, config); err != nil {
		return err
	}
	username, ok := b.consumeResetToken(req.ResetToken)
	if !ok {
		return errors.New("密码重置凭证无效或已过期")
	}
	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		return err
	}
	result := b.db.Model(&model.User{}).Where("username = ? AND status = ?", username, true).Update("password_hash", hash)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrInvalidCredentials
	}
	return nil
}

func (b *AuthBiz) loginVerificationRequired(config *api.SystemConfigResponse, user *model.User) bool {
	if !config.LoginVerificationEnabled {
		return false
	}
	if config.LoginVerificationMethod == "email" {
		return config.VerificationEmailEnabled && strings.TrimSpace(user.Email) != ""
	}
	return config.LoginVerificationMethod == "webhook" && config.VerificationWebhookEnabled
}

func (b *AuthBiz) authenticate(username, password string) (*model.User, error) {
	user, err := b.findLoginUser(username)
	if err != nil || !auth.CheckPassword(user.PasswordHash, password) {
		return nil, ErrInvalidCredentials
	}
	return user, nil
}

func (b *AuthBiz) findLoginUser(identifier string) (*model.User, error) {
	var user model.User
	identifier = strings.TrimSpace(identifier)
	var err error
	if strings.Contains(identifier, "@") {
		err = b.db.Where("email <> '' AND LOWER(email) = ?", strings.ToLower(identifier)).First(&user).Error
	} else {
		err = b.db.Where("username = ?", identifier).First(&user).Error
	}
	if err != nil || !user.Status {
		return nil, ErrInvalidCredentials
	}
	return &user, nil
}

func (b *AuthBiz) verifyCaptcha(id, code string, consume bool) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	value, ok := b.captchas[id]
	if !ok || time.Now().After(value.expiresAt) || !strings.EqualFold(value.code, strings.TrimSpace(code)) {
		delete(b.captchas, id)
		return false
	}
	if consume {
		delete(b.captchas, id)
	}
	return true
}

func (b *AuthBiz) verifyEmailCode(username, code string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	value, ok := b.emailCodes[username]
	if !ok {
		return false
	}
	if time.Now().After(value.expiresAt) {
		if time.Now().After(value.challengeExpiresAt) {
			delete(b.emailCodes, username)
		}
		return false
	}
	if value.code != strings.TrimSpace(code) {
		value.attempts++
		if value.attempts >= 5 {
			delete(b.emailCodes, username)
		} else {
			b.emailCodes[username] = value
		}
		return false
	}
	delete(b.emailCodes, username)
	return true
}

func (b *AuthBiz) hasActiveEmailCode(username string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	value, ok := b.emailCodes[username]
	if ok && time.Now().Before(value.challengeExpiresAt) {
		return true
	}
	delete(b.emailCodes, username)
	return false
}

func (b *AuthBiz) reserveEmailCode(username, code string, expiresIn time.Duration) (int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	value, ok := b.emailCodes[username]
	if !ok || now.After(value.expiresAt) {
		delete(b.emailCodes, username)
	} else if remaining := value.expiresAt.Sub(now); remaining > 0 {
		return int(remaining.Seconds()) + 1, int(remaining.Seconds()) + 1
	}
	b.emailCodes[username] = verificationValue{code: code, expiresAt: now.Add(expiresIn), challengeExpiresAt: now.Add(expiresIn + 30*time.Minute)}
	return 0, int(expiresIn.Seconds())
}

func (b *AuthBiz) discardEmailCode(username, code string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if value, ok := b.emailCodes[username]; ok && value.code == code {
		delete(b.emailCodes, username)
	}
}

func (b *AuthBiz) reserveResetCode(username, code string, expiresIn time.Duration) (int, int) {
	b.mu.Lock()
	defer b.mu.Unlock()
	now := time.Now()
	value, ok := b.resetCodes[username]
	if ok && now.Before(value.expiresAt) {
		remaining := value.expiresAt.Sub(now)
		return int(remaining.Seconds()) + 1, int(remaining.Seconds()) + 1
	}
	b.resetCodes[username] = verificationValue{code: code, expiresAt: now.Add(expiresIn)}
	return 0, int(expiresIn.Seconds())
}

func (b *AuthBiz) discardResetCode(username, code string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if value, ok := b.resetCodes[username]; ok && value.code == code {
		delete(b.resetCodes, username)
	}
}

func (b *AuthBiz) verifyResetCode(username, code string) bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	value, ok := b.resetCodes[username]
	if !ok || time.Now().After(value.expiresAt) {
		delete(b.resetCodes, username)
		return false
	}
	if value.code != strings.TrimSpace(code) {
		value.attempts++
		if value.attempts >= 5 {
			delete(b.resetCodes, username)
		} else {
			b.resetCodes[username] = value
		}
		return false
	}
	delete(b.resetCodes, username)
	return true
}

func (b *AuthBiz) consumeResetToken(token string) (string, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	value, ok := b.resetTokens[token]
	delete(b.resetTokens, token)
	return value.code, ok && time.Now().Before(value.expiresAt)
}

func randomToken() (string, error) {
	buffer := make([]byte, 16)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return hex.EncodeToString(buffer), nil
}

func randomDigits(length int) (string, error) {
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		result[i] = byte('0' + n.Int64())
	}
	return string(result), nil
}

func (b *AuthBiz) RecordInvalidLogin(username, ip, message string) {
	b.recordLogin(0, username, ip, "failed", message)
}

func (b *AuthBiz) recordLogin(userID int64, username, ip, status, message string) {
	log := &model.AuditLoginLog{UserID: userID, Username: username, LoginType: "web", Status: status, IP: ip, Message: message}
	if err := b.db.Create(log).Error; err != nil {
		slog.Error("保存登录日志失败", "error", err, "username", username, "ip", ip)
	}
}

func (b *AuthBiz) Me(userID uint) (*api.CurrentUserResponse, error) {
	var user model.User
	if err := b.db.Preload("Roles").Preload("Organizations").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	return &api.CurrentUserResponse{User: toUserResponse(user), Permissions: b.casbin.AllowedPermissionKeys(userID)}, nil
}

func (b *AuthBiz) MFAStatus(userID uint) (*api.MFAStatusResponse, error) {
	var user model.User
	if err := b.db.Select("id", "email").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	var totp model.UserMfaTotp
	totpConfigured := b.db.Where("user_id = ?", userID).First(&totp).Error == nil
	totpEnabled := totpConfigured && totp.Enabled == 1
	var passkeys []model.UserMfaPasskey
	if err := b.db.Where("user_id = ?", userID).Order("id DESC").Find(&passkeys).Error; err != nil {
		return nil, err
	}
	items := make([]*api.MFAPasskeyResponse, 0, len(passkeys))
	for _, passkey := range passkeys {
		items = append(items, &api.MFAPasskeyResponse{ID: passkey.ID, Name: passkey.Name, MFAEnabled: passkey.Enabled == 1, DirectLoginEnabled: passkey.DirectLoginEnabled == 1})
	}
	var emailMFA model.UserMfaEmail
	_ = b.db.Where("user_id = ?", userID).First(&emailMFA).Error
	emailAvailable := config.VerificationEmailEnabled && strings.TrimSpace(user.Email) != ""
	return &api.MFAStatusResponse{CodeAvailable: passwordResetVerificationMethod(config, &user) != "", EmailAvailable: emailAvailable, Email: user.Email, EmailMFAEnabled: emailAvailable && emailMFA.MfaEnabled == 1, EmailDirectLoginEnabled: emailAvailable && emailMFA.DirectLoginEnabled == 1, TOTPConfigured: totpConfigured, TOTPEnabled: totpEnabled, Passkeys: items}, nil
}

func (b *AuthBiz) BeginTOTPSetup(userID uint) (*api.TOTPSetupResponse, error) {
	var user model.User
	if err := b.db.Select("id", "username").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	secret, err := auth.GenerateTOTPSecret()
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.pendingTOTP[userID] = verificationValue{code: secret, expiresAt: time.Now().Add(10 * time.Minute)}
	b.mu.Unlock()
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	uri := auth.TOTPURI(config.SystemName, user.Username, secret)
	image, err := qrcode.Encode(uri, qrcode.Medium, 220)
	if err != nil {
		return nil, err
	}
	return &api.TOTPSetupResponse{Secret: secret, URI: uri, QRCode: "data:image/png;base64," + base64.StdEncoding.EncodeToString(image)}, nil
}

func (b *AuthBiz) ConfirmTOTPSetup(userID uint, req *api.TOTPConfirmRequest) error {
	b.mu.Lock()
	pending, ok := b.pendingTOTP[userID]
	if ok && time.Now().After(pending.expiresAt) {
		delete(b.pendingTOTP, userID)
		ok = false
	}
	b.mu.Unlock()
	if !ok || !auth.ValidateTOTP(pending.code, req.Code, time.Now()) {
		return errors.New("TOTP 验证码错误或设置已过期")
	}
	encrypted, err := cryptox.Encrypt(b.config.aesKey, pending.code)
	if err != nil {
		return err
	}
	row := &model.UserMfaTotp{UserID: int64(userID), Secret: encrypted, Enabled: 1}
	if err := b.db.Where("user_id = ?", userID).Assign(map[string]any{"secret": encrypted, "enabled": 1, "updated_at": time.Now()}).FirstOrCreate(row).Error; err != nil {
		return err
	}
	b.mu.Lock()
	delete(b.pendingTOTP, userID)
	b.mu.Unlock()
	return nil
}

func (b *AuthBiz) DisableTOTP(userID uint, req *api.TOTPConfirmRequest) error {
	var row model.UserMfaTotp
	if err := b.db.Where("user_id = ? AND enabled = ?", userID, 1).First(&row).Error; err != nil {
		return ErrNotFound
	}
	secret, err := cryptox.Decrypt(b.config.aesKey, row.Secret)
	if err != nil {
		return err
	}
	if !auth.ValidateTOTP(secret, req.Code, time.Now()) {
		return errors.New("TOTP 验证码错误")
	}
	return b.db.Model(&row).Updates(map[string]any{"enabled": 0, "updated_at": time.Now()}).Error
}

func (b *AuthBiz) EnableTOTP(userID uint, req *api.TOTPConfirmRequest) error {
	var row model.UserMfaTotp
	if err := b.db.Where("user_id = ? AND enabled = ?", userID, 0).First(&row).Error; err != nil {
		return ErrNotFound
	}
	secret, err := cryptox.Decrypt(b.config.aesKey, row.Secret)
	if err != nil {
		return err
	}
	if !auth.ValidateTOTP(secret, req.Code, time.Now()) {
		return errors.New("TOTP 验证码错误")
	}
	return b.db.Model(&row).Updates(map[string]any{"enabled": 1, "updated_at": time.Now()}).Error
}

func (b *AuthBiz) DeleteTOTP(userID uint, req *api.TOTPConfirmRequest) error {
	var row model.UserMfaTotp
	if err := b.db.Where("user_id = ?", userID).First(&row).Error; err != nil {
		return ErrNotFound
	}
	secret, err := cryptox.Decrypt(b.config.aesKey, row.Secret)
	if err != nil {
		return err
	}
	if !auth.ValidateTOTP(secret, req.Code, time.Now()) {
		return errors.New("TOTP 验证码错误")
	}
	return b.db.Delete(&row).Error
}

func (b *AuthBiz) UpdateProfile(userID uint, req *api.ProfileUpdateRequest) (*api.UserResponse, error) {
	var user model.User
	if err := b.db.Preload("Roles").Preload("Organizations").First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	username := strings.TrimSpace(req.Username)
	email := strings.ToLower(strings.TrimSpace(req.Email))
	nickname := strings.TrimSpace(req.Nickname)
	if username == "" || nickname == "" {
		return nil, errors.New("用户名和昵称不能为空")
	}
	var duplicateCount int64
	if err := b.db.Model(&model.User{}).Where("id <> ? AND username = ?", user.ID, username).Count(&duplicateCount).Error; err != nil {
		return nil, err
	}
	if duplicateCount > 0 {
		return nil, errors.New("用户名已存在")
	}
	if email != "" {
		if err := b.db.Model(&model.User{}).Where("id <> ? AND email = ?", user.ID, email).Count(&duplicateCount).Error; err != nil {
			return nil, err
		}
		if duplicateCount > 0 {
			return nil, errors.New("邮箱已被使用")
		}
	}
	if req.NewPassword != "" {
		if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
			return nil, errors.New("当前密码不正确")
		}
		config, err := b.config.Get()
		if err != nil {
			return nil, err
		}
		if err := validateProfilePassword(req.NewPassword, config); err != nil {
			return nil, err
		}
		hash, err := auth.HashPassword(req.NewPassword)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	user.Username = username
	user.Nickname = nickname
	user.Email = email
	if err := b.db.Save(&user).Error; err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func validateProfilePassword(password string, config *api.SystemConfigResponse) error {
	if len([]rune(password)) < config.PasswordMinLength {
		return fmt.Errorf("新密码长度不能少于 %d 个字符", config.PasswordMinLength)
	}
	var hasDigit, hasLetter, hasUpper, hasLower, hasSpecial bool
	for _, character := range password {
		hasDigit = hasDigit || unicode.IsDigit(character)
		hasLetter = hasLetter || unicode.IsLetter(character)
		hasUpper = hasUpper || unicode.IsUpper(character)
		hasLower = hasLower || unicode.IsLower(character)
		hasSpecial = hasSpecial || (!unicode.IsLetter(character) && !unicode.IsDigit(character))
	}
	if config.PasswordRequireDigit && !hasDigit || config.PasswordRequireLetter && !hasLetter ||
		config.PasswordRequireUpper && !hasUpper || config.PasswordRequireLower && !hasLower ||
		config.PasswordRequireSpecial && !hasSpecial {
		return errors.New("新密码不符合系统密码安全策略")
	}
	return nil
}
