package biz

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/cryptox"

	"github.com/go-webauthn/webauthn/webauthn"
)

type webAuthnUser struct {
	id          uint
	name        string
	displayName string
	credentials []webauthn.Credential
}

func (u *webAuthnUser) WebAuthnID() []byte                         { return []byte(fmt.Sprintf("gops-user-%d", u.id)) }
func (u *webAuthnUser) WebAuthnName() string                       { return u.name }
func (u *webAuthnUser) WebAuthnDisplayName() string                { return u.displayName }
func (u *webAuthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

type passkeySession struct {
	userID    uint
	username  string
	ip        string
	name      string
	origin    string
	session   *webauthn.SessionData
	expiresAt time.Time
}

func (b *AuthBiz) getMFAChallenge(token string) (loginMFAChallenge, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	challenge, ok := b.mfaChallenges[token]
	if ok && time.Now().After(challenge.expiresAt) {
		delete(b.mfaChallenges, token)
		return loginMFAChallenge{}, false
	}
	return challenge, ok
}

func containsMethod(methods []string, method string) bool {
	for _, item := range methods {
		if item == method {
			return true
		}
	}
	return false
}

func (b *AuthBiz) SendMFACode(req *api.MFACodeSendRequest) (*api.EmailCodeResponse, error) {
	challenge, ok := b.getMFAChallenge(req.MFAToken)
	if !ok || !containsMethod(challenge.methods, "code") {
		return nil, errors.New("MFA 登录会话无效或已过期")
	}
	var user model.User
	if err := b.db.First(&user, challenge.userID).Error; err != nil {
		return nil, ErrInvalidCredentials
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	expiresIn := 5 * time.Minute
	if config.LoginVerificationMethod == "email" {
		expiresIn = time.Duration(config.VerificationEmailExpireMinutes) * time.Minute
	}
	code, err := randomDigits(6)
	if err != nil {
		return nil, err
	}
	key := "mfa:" + req.MFAToken
	if retryAfter, expiresAfter := b.reserveEmailCode(key, code, expiresIn); retryAfter > 0 {
		return &api.EmailCodeResponse{Required: true, Sent: false, RetryAfterSeconds: retryAfter, ExpiresInSeconds: expiresAfter}, nil
	}
	go func() {
		if sendErr := b.config.SendLoginVerification(config, user.Username, user.Email, code); sendErr != nil {
			b.discardEmailCode(key, code)
		}
	}()
	return &api.EmailCodeResponse{Required: true, Sent: true, RetryAfterSeconds: int(expiresIn.Seconds()), ExpiresInSeconds: int(expiresIn.Seconds())}, nil
}

func (b *AuthBiz) VerifyMFA(req *api.MFAVerifyRequest) (*api.LoginResponse, error) {
	challenge, ok := b.getMFAChallenge(req.MFAToken)
	if !ok || !containsMethod(challenge.methods, req.Method) {
		return nil, errors.New("MFA 登录会话无效或已过期")
	}
	valid := false
	switch req.Method {
	case "code":
		valid = b.verifyEmailCode("mfa:"+req.MFAToken, req.Code)
	case "totp":
		var row model.UserMfaTotp
		if err := b.db.Where("user_id = ? AND enabled = ?", challenge.userID, 1).First(&row).Error; err == nil {
			secret, decryptErr := cryptox.Decrypt(b.config.aesKey, row.Secret)
			valid = decryptErr == nil && auth.ValidateTOTP(secret, req.Code, time.Now())
		}
	default:
		return nil, errors.New("不支持的 MFA 验证方式")
	}
	if !valid {
		b.recordLogin(int64(challenge.userID), challenge.username, challenge.ip, "failed", "MFA 验证失败")
		return nil, errors.New("验证码错误或已过期")
	}
	return b.finishMFALogin(req.MFAToken, challenge)
}

func (b *AuthBiz) finishMFALogin(token string, challenge loginMFAChallenge) (*api.LoginResponse, error) {
	tokens, err := b.jwt.GeneratePair(challenge.userID)
	if err != nil {
		return nil, ErrGenerateToken
	}
	b.mu.Lock()
	delete(b.mfaChallenges, token)
	b.mu.Unlock()
	b.recordLogin(int64(challenge.userID), challenge.username, challenge.ip, "success", "MFA 登录成功")
	return tokenResponse(tokens), nil
}

func webAuthnForOrigin(origin, displayName string) (*webauthn.WebAuthn, error) {
	parsed, err := url.Parse(origin)
	if err != nil || parsed.Scheme == "" || parsed.Hostname() == "" {
		return nil, errors.New("无效的 WebAuthn 来源")
	}
	return webauthn.New(&webauthn.Config{RPID: parsed.Hostname(), RPDisplayName: displayName, RPOrigins: []string{parsed.Scheme + "://" + parsed.Host}})
}

func (b *AuthBiz) loadWebAuthnUser(userID uint, usage string) (*webAuthnUser, error) {
	var user model.User
	if err := b.db.First(&user, userID).Error; err != nil {
		return nil, ErrNotFound
	}
	var rows []model.UserMfaPasskey
	query := b.db.Where("user_id = ?", userID)
	if usage == "mfa" {
		query = query.Where("enabled = ?", 1)
	} else if usage == "direct" {
		query = query.Where("direct_login_enabled = ?", 1)
	}
	if err := query.Find(&rows).Error; err != nil {
		return nil, err
	}
	credentials := make([]webauthn.Credential, 0, len(rows))
	for _, row := range rows {
		plain, err := cryptox.Decrypt(b.config.aesKey, row.CredentialJSON)
		if err != nil {
			return nil, err
		}
		var credential webauthn.Credential
		if err := json.Unmarshal([]byte(plain), &credential); err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	return &webAuthnUser{id: userID, name: user.Username, displayName: user.Nickname, credentials: credentials}, nil
}

func credentialID(value []byte) string { return base64.RawURLEncoding.EncodeToString(value) }

func credentialRequest(origin string, payload []byte) *http.Request {
	r := httptest.NewRequest(http.MethodPost, origin, strings.NewReader(string(payload)))
	r.Header.Set("Content-Type", "application/json")
	r.Header.Set("Origin", origin)
	return r
}

func (b *AuthBiz) BeginPasskeyRegistration(userID uint, name, origin string) (*api.PasskeyRegisterBeginResponse, error) {
	user, err := b.loadWebAuthnUser(userID, "")
	if err != nil {
		return nil, err
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	wa, err := webAuthnForOrigin(origin, config.SystemName)
	if err != nil {
		return nil, err
	}
	options, session, err := wa.BeginRegistration(user)
	if err != nil {
		return nil, err
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.passkeyRegistrations[token] = passkeySession{userID: userID, name: strings.TrimSpace(name), origin: origin, session: session, expiresAt: time.Now().Add(5 * time.Minute)}
	b.mu.Unlock()
	return &api.PasskeyRegisterBeginResponse{RegistrationToken: token, Options: options}, nil
}

func (b *AuthBiz) FinishPasskeyRegistration(userID uint, req *api.PasskeyRegisterFinishRequest) error {
	b.mu.Lock()
	session, ok := b.passkeyRegistrations[req.RegistrationToken]
	delete(b.passkeyRegistrations, req.RegistrationToken)
	b.mu.Unlock()
	if !ok || session.userID != userID || time.Now().After(session.expiresAt) {
		return errors.New("Passkey 注册会话无效或已过期")
	}
	user, err := b.loadWebAuthnUser(userID, "")
	if err != nil {
		return err
	}
	config, err := b.config.Get()
	if err != nil {
		return err
	}
	wa, err := webAuthnForOrigin(session.origin, config.SystemName)
	if err != nil {
		return err
	}
	credential, err := wa.FinishRegistration(user, *session.session, credentialRequest(session.origin, req.Credential))
	if err != nil {
		return errors.New("Passkey 注册验证失败")
	}
	encoded, err := json.Marshal(credential)
	if err != nil {
		return err
	}
	encrypted, err := cryptox.Encrypt(b.config.aesKey, string(encoded))
	if err != nil {
		return err
	}
	row := &model.UserMfaPasskey{UserID: int64(userID), Name: strings.TrimSpace(req.Name), CredentialID: credentialID(credential.ID), CredentialJSON: encrypted, Enabled: 1}
	return b.db.Create(row).Error
}

func (b *AuthBiz) DeletePasskey(userID uint, req *api.PasskeyDeleteRequest) error {
	result := b.db.Where("id = ? AND user_id = ?", req.ID, userID).Delete(&model.UserMfaPasskey{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *AuthBiz) UpdatePasskeyStatus(userID uint, req *api.PasskeyStatusRequest) error {
	result := b.db.Model(&model.UserMfaPasskey{}).Where("id = ? AND user_id = ?", req.ID, userID).Updates(map[string]any{"enabled": req.MFAEnabled, "direct_login_enabled": req.DirectLoginEnabled, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *AuthBiz) RenamePasskey(userID uint, req *api.PasskeyRenameRequest) error {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return errors.New("Passkey 名称不能为空")
	}
	if len([]rune(name)) > 64 {
		return errors.New("Passkey 名称不能超过 64 个字符")
	}
	result := b.db.Model(&model.UserMfaPasskey{}).Where("id = ? AND user_id = ?", req.ID, userID).Updates(map[string]any{"name": name, "updated_at": time.Now()})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func (b *AuthBiz) BeginPasskeyLogin(req *api.PasskeyLoginBeginRequest, origin string) (*api.PasskeyLoginBeginResponse, error) {
	challenge, ok := b.getMFAChallenge(req.MFAToken)
	if !ok || !containsMethod(challenge.methods, "passkey") {
		return nil, errors.New("MFA 登录会话无效或已过期")
	}
	user, err := b.loadWebAuthnUser(challenge.userID, "mfa")
	if err != nil {
		return nil, err
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	wa, err := webAuthnForOrigin(origin, config.SystemName)
	if err != nil {
		return nil, err
	}
	options, session, err := wa.BeginLogin(user)
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.passkeyLogins[req.MFAToken] = passkeySession{userID: challenge.userID, origin: origin, session: session, expiresAt: time.Now().Add(5 * time.Minute)}
	b.mu.Unlock()
	return &api.PasskeyLoginBeginResponse{Options: options}, nil
}

func (b *AuthBiz) FinishPasskeyLogin(req *api.PasskeyLoginFinishRequest) (*api.LoginResponse, error) {
	challenge, ok := b.getMFAChallenge(req.MFAToken)
	if !ok || !containsMethod(challenge.methods, "passkey") {
		return nil, errors.New("MFA 登录会话无效或已过期")
	}
	b.mu.Lock()
	session, sessionOK := b.passkeyLogins[req.MFAToken]
	delete(b.passkeyLogins, req.MFAToken)
	b.mu.Unlock()
	if !sessionOK || time.Now().After(session.expiresAt) {
		return nil, errors.New("Passkey 验证会话无效或已过期")
	}
	user, err := b.loadWebAuthnUser(challenge.userID, "mfa")
	if err != nil {
		return nil, err
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	wa, err := webAuthnForOrigin(session.origin, config.SystemName)
	if err != nil {
		return nil, err
	}
	credential, err := wa.FinishLogin(user, *session.session, credentialRequest(session.origin, req.Credential))
	if err != nil {
		return nil, errors.New("Passkey 验证失败")
	}
	encoded, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	encrypted, err := cryptox.Encrypt(b.config.aesKey, string(encoded))
	if err != nil {
		return nil, err
	}
	if err := b.db.Model(&model.UserMfaPasskey{}).Where("user_id = ? AND credential_id = ?", challenge.userID, credentialID(credential.ID)).Updates(map[string]any{"credential_json": encrypted, "updated_at": time.Now()}).Error; err != nil {
		return nil, err
	}
	return b.finishMFALogin(req.MFAToken, challenge)
}

func (b *AuthBiz) BeginDirectPasskeyLogin(req *api.DirectPasskeyLoginBeginRequest, origin, ip string) (*api.DirectPasskeyLoginBeginResponse, error) {
	userModel, err := b.findLoginUser(req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	if config.CaptchaEnabled && !b.verifyCaptcha(req.CaptchaID, req.CaptchaCode, true) {
		return nil, ErrInvalidCaptcha
	}
	user, err := b.loadWebAuthnUser(uint(userModel.ID), "direct")
	if err != nil {
		return nil, err
	}
	if len(user.credentials) == 0 {
		return nil, errors.New("当前用户没有启用可直接登录的 Passkey")
	}
	wa, err := webAuthnForOrigin(origin, config.SystemName)
	if err != nil {
		return nil, err
	}
	options, session, err := wa.BeginLogin(user)
	if err != nil {
		return nil, err
	}
	token, err := randomToken()
	if err != nil {
		return nil, err
	}
	b.mu.Lock()
	b.directPasskeyLogins[token] = passkeySession{userID: uint(userModel.ID), username: userModel.Username, ip: ip, origin: origin, session: session, expiresAt: time.Now().Add(5 * time.Minute)}
	b.mu.Unlock()
	return &api.DirectPasskeyLoginBeginResponse{LoginToken: token, Options: options}, nil
}

func (b *AuthBiz) FinishDirectPasskeyLogin(req *api.DirectPasskeyLoginFinishRequest) (*api.LoginResponse, error) {
	b.mu.Lock()
	session, ok := b.directPasskeyLogins[req.LoginToken]
	delete(b.directPasskeyLogins, req.LoginToken)
	b.mu.Unlock()
	if !ok || time.Now().After(session.expiresAt) {
		return nil, errors.New("Passkey 登录会话无效或已过期")
	}
	user, err := b.loadWebAuthnUser(session.userID, "direct")
	if err != nil {
		return nil, err
	}
	config, err := b.config.Get()
	if err != nil {
		return nil, err
	}
	wa, err := webAuthnForOrigin(session.origin, config.SystemName)
	if err != nil {
		return nil, err
	}
	credential, err := wa.FinishLogin(user, *session.session, credentialRequest(session.origin, req.Credential))
	if err != nil {
		b.recordLogin(int64(session.userID), session.username, session.ip, "failed", "Passkey 直接登录失败")
		return nil, errors.New("Passkey 验证失败")
	}
	encoded, err := json.Marshal(credential)
	if err != nil {
		return nil, err
	}
	encrypted, err := cryptox.Encrypt(b.config.aesKey, string(encoded))
	if err != nil {
		return nil, err
	}
	if err := b.db.Model(&model.UserMfaPasskey{}).Where("user_id = ? AND credential_id = ? AND direct_login_enabled = ?", session.userID, credentialID(credential.ID), 1).Updates(map[string]any{"credential_json": encrypted, "updated_at": time.Now()}).Error; err != nil {
		return nil, err
	}
	tokens, err := b.jwt.GeneratePair(session.userID)
	if err != nil {
		return nil, ErrGenerateToken
	}
	b.recordLogin(int64(session.userID), session.username, session.ip, "success", "Passkey 直接登录成功")
	return tokenResponse(tokens), nil
}
