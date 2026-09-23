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

type AuthBiz struct {
	db         *gorm.DB
	jwt        *auth.JWTManager
	casbin     *casbinx.Service
	config     *SystemConfigBiz
	mu         sync.Mutex
	captchas   map[string]verificationValue
	emailCodes map[string]verificationValue
}

func NewAuthBiz(db *gorm.DB, jwt *auth.JWTManager, casbinService *casbinx.Service, config *SystemConfigBiz) *AuthBiz {
	return &AuthBiz{db: db, jwt: jwt, casbin: casbinService, config: config, captchas: make(map[string]verificationValue), emailCodes: make(map[string]verificationValue)}
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
	verificationRequired := b.loginVerificationRequired(config, user)
	if config.CaptchaEnabled && !verificationRequired && !b.verifyCaptcha(req.CaptchaID, req.CaptchaCode, true) {
		b.recordLogin(user.ID, user.Username, ip, "failed", "图形验证码错误或已过期")
		return nil, ErrInvalidCaptcha
	}
	if verificationRequired && !b.verifyEmailCode(user.Username, req.EmailCode) {
		b.recordLogin(user.ID, user.Username, ip, "failed", "登录验证码错误或已过期")
		return nil, ErrInvalidEmailCode
	}

	token, err := b.jwt.Generate(uint(user.ID))
	if err != nil {
		b.recordLogin(user.ID, user.Username, ip, "failed", "生成登录令牌失败")
		return nil, ErrGenerateToken
	}
	b.recordLogin(user.ID, user.Username, ip, "success", "登录成功")
	return &api.LoginResponse{Token: token}, nil
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
			slog.Error("发送登录二次验证码失败", "error", err, "username", username, "method", config.LoginVerificationMethod)
		}
	}(user.Username, user.Email, code)
	return &api.EmailCodeResponse{Required: true, Sent: true, RetryAfterSeconds: int(expiresIn.Seconds()), ExpiresInSeconds: int(expiresIn.Seconds())}, nil
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
	var user model.User
	if err := b.db.Where("username = ?", username).First(&user).Error; err != nil || !user.Status || !auth.CheckPassword(user.PasswordHash, password) {
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
