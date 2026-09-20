package biz

import (
	"context"
	"crypto/tls"
	"crypto/x509/pkix"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/cryptox"
	"gorm.io/gorm"
)

type MonitorBiz struct {
	db     *gorm.DB
	client *http.Client
	aesKey string
}

func NewMonitorBiz(db *gorm.DB, aesKey string) *MonitorBiz {
	return &MonitorBiz{db: db, client: &http.Client{Timeout: 15 * time.Second}, aesKey: aesKey}
}

// StartScheduler periodically checks monitor entries whose next check time has arrived.
func (b *MonitorBiz) StartScheduler(ctx context.Context) {
	b.CheckDueDomains()
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			b.CheckDueDomains()
		}
	}
}

// CheckDueDomains executes due checks in background without blocking the scheduler.
func (b *MonitorBiz) CheckDueDomains() {
	var rows []model.MonitorDomain
	now := time.Now()
	if err := b.db.Where("status <> ? AND (next_check IS NULL OR next_check <= ?)", "paused", now).Find(&rows).Error; err != nil {
		return
	}
	for _, row := range rows {
		row := row
		go func() { _ = b.checkDomain(&row) }()
	}
}

func (b *MonitorBiz) ListDomains(req *api.MonitorDomainListRequest) ([]*api.MonitorDomainResponse, error) {
	var rows []model.MonitorDomain
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("domain LIKE ?", "%"+req.Keyword+"%")
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*api.MonitorDomainResponse, 0, len(rows))
	for _, row := range rows {
		out = append(out, monitorDomainResponse(row))
	}
	return out, nil
}
func (b *MonitorBiz) CreateDomain(req *api.MonitorDomainRequest) (*api.MonitorDomainResponse, error) {
	normalized := normalizeMonitorDomainRequest(req)
	row := b.domainModel(normalized)
	row.NextCheck = time.Now()
	if err := b.db.Create(&row).Error; err != nil {
		return nil, err
	}
	return monitorDomainResponse(row), nil
}
func (b *MonitorBiz) UpdateDomain(req *api.MonitorDomainRequest) (*api.MonitorDomainResponse, error) {
	var row model.MonitorDomain
	if err := b.db.First(&row, req.ID).Error; err != nil {
		return nil, monitorError(err)
	}
	normalized := normalizeMonitorDomainRequest(req)
	updated := b.domainModel(normalized)
	updated.ID, updated.CreatedAt = row.ID, row.CreatedAt
	if err := b.db.Save(&updated).Error; err != nil {
		return nil, err
	}
	return monitorDomainResponse(updated), nil
}
func (b *MonitorBiz) DeleteDomain(req *api.MonitorDomainIDRequest) error {
	r := b.db.Delete(&model.MonitorDomain{}, req.ID)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *MonitorBiz) CheckDomain(req *api.MonitorDomainIDRequest) (*api.MonitorDomainResponse, error) {
	var row model.MonitorDomain
	if err := b.db.First(&row, req.ID).Error; err != nil {
		return nil, monitorError(err)
	}
	if err := b.checkDomain(&row); err != nil {
		return nil, err
	}
	return monitorDomainResponse(row), nil
}

func (b *MonitorBiz) checkDomain(row *model.MonitorDomain) error {
	normalized := normalizeMonitorDomainAddress(row.Domain, row.Port, row.EnableSsl)
	row.Domain, row.Port, row.EnableSsl = normalized.domain, normalized.port, normalized.enableSSL
	if !row.EnableSsl {
		row.VerifyCertificate = false
	}
	start := time.Now()
	scheme := "http"
	if row.EnableSsl {
		scheme = "https"
	}
	client := b.client
	if row.EnableSsl && !row.VerifyCertificate {
		transport := http.DefaultTransport.(*http.Transport).Clone()
		transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} // #nosec G402 -- explicitly configured to skip certificate verification.
		client = &http.Client{Timeout: b.client.Timeout, Transport: transport}
	}
	res, err := client.Get(fmt.Sprintf("%s://%s:%d", scheme, row.Domain, row.Port))
	now := time.Now()
	row.LastCheck, row.ResponseTime, row.NextCheck = now, time.Since(start).Milliseconds(), now.Add(time.Duration(row.CheckInterval)*time.Second)
	row.LastError, row.Status = "", "normal"
	statusCode := int64(0)
	if err != nil {
		row.Status, row.LastError = "abnormal", err.Error()
	} else {
		_ = res.Body.Close()
		statusCode = int64(res.StatusCode)
		if res.StatusCode >= 400 {
			row.Status, row.LastError = "abnormal", res.Status
		}
	}
	row.SslValid = false
	row.SslExpiry = time.Time{}
	row.SslRootCa = ""
	if row.EnableSsl {
		conn, tlsErr := tls.Dial("tcp", fmt.Sprintf("%s:%d", row.Domain, row.Port), &tls.Config{InsecureSkipVerify: !row.VerifyCertificate}) // #nosec G402 -- controlled by the monitor certificate-verification setting.
		if tlsErr != nil {
			if row.Status == "normal" {
				row.Status, row.LastError = "abnormal", tlsErr.Error()
			}
		} else {
			state := conn.ConnectionState()
			_ = conn.Close()
			if len(state.PeerCertificates) > 0 {
				row.SslExpiry = state.PeerCertificates[0].NotAfter
				row.SslValid = time.Now().Before(row.SslExpiry)
				row.SslRootCa = rootCAName(state)
			}
		}
	}
	if err := b.db.Save(row).Error; err != nil {
		return err
	}
	if err := b.db.Create(&model.MonitorCheckHistory{DomainID: row.ID, Status: row.Status, ResponseTime: row.ResponseTime, SslValid: row.SslValid, SslExpiry: row.SslExpiry, StatusCode: statusCode, ErrorMessage: row.LastError, CheckedAt: now}).Error; err != nil {
		return err
	}
	if err := b.pruneCheckHistory(row.ID); err != nil {
		return err
	}
	if row.Status == "abnormal" && row.EnableAlert {
		_ = b.db.Create(&model.MonitorAlertLog{DomainID: row.ID, AlertType: "availability", Status: "pending", Message: row.LastError, SentAt: time.Now()}).Error
	}
	if row.EnableAlert && row.ResponseThreshold > 0 && row.ResponseTime > row.ResponseThreshold {
		_ = b.db.Create(&model.MonitorAlertLog{DomainID: row.ID, AlertType: "response_time", Status: "pending", Message: fmt.Sprintf("响应时间 %dms 超过阈值 %dms", row.ResponseTime, row.ResponseThreshold), SentAt: time.Now()}).Error
	}
	if row.EnableAlert && row.EnableSsl && !row.SslExpiry.IsZero() && time.Until(row.SslExpiry) <= time.Duration(row.SslExpiryDays)*24*time.Hour {
		_ = b.db.Create(&model.MonitorAlertLog{DomainID: row.ID, AlertType: "ssl_expiry", Status: "pending", Message: fmt.Sprintf("SSL 证书将于 %s 到期", row.SslExpiry.Format("2006-01-02 15:04:05")), SentAt: time.Now()}).Error
	}
	return nil
}
func (b *MonitorBiz) DomainStats() (*api.MonitorStatsResponse, error) {
	out := &api.MonitorStatsResponse{}
	if err := b.db.Model(&model.MonitorDomain{}).Count(&out.Total).Error; err != nil {
		return nil, err
	}
	for _, item := range []struct {
		status string
		into   *int64
	}{{"normal", &out.Normal}, {"abnormal", &out.Abnormal}, {"paused", &out.Paused}} {
		if err := b.db.Model(&model.MonitorDomain{}).Where("status = ?", item.status).Count(item.into).Error; err != nil {
			return nil, err
		}
	}
	return out, nil
}

// ListCheckHistory returns recent request logs for one domain monitor.
func (b *MonitorBiz) ListCheckHistory(req *api.MonitorCheckHistoryListRequest) ([]*api.MonitorCheckHistoryResponse, error) {
	if err := b.pruneCheckHistory(req.DomainID); err != nil {
		return nil, err
	}
	var rows []model.MonitorCheckHistory
	if err := b.db.Where("domain_id = ?", req.DomainID).Order("checked_at DESC, id DESC").Limit(30).Find(&rows).Error; err != nil {
		return nil, err
	}
	result := make([]*api.MonitorCheckHistoryResponse, 0, len(rows))
	for _, row := range rows {
		result = append(result, &api.MonitorCheckHistoryResponse{ID: row.ID, DomainID: row.DomainID, Status: row.Status, ResponseTime: row.ResponseTime, SSLValid: row.SslValid, SSLExpiry: row.SslExpiry, StatusCode: row.StatusCode, ErrorMessage: row.ErrorMessage, CheckedAt: row.CheckedAt})
	}
	return result, nil
}

func (b *MonitorBiz) pruneCheckHistory(domainID int64) error {
	keep := b.db.Model(&model.MonitorCheckHistory{}).
		Select("id").
		Where("domain_id = ?", domainID).
		Order("checked_at DESC, id DESC").
		Limit(30)
	return b.db.Where("domain_id = ? AND id NOT IN (?)", domainID, keep).Delete(&model.MonitorCheckHistory{}).Error
}
func (b *MonitorBiz) domainModel(req *api.MonitorDomainRequest) model.MonitorDomain {
	p, i, d := req.Port, req.CheckInterval, req.SSLExpiryDays
	if p == 0 {
		p = 443
	}
	if i == 0 {
		i = 300
	}
	if d == 0 {
		d = 30
	}
	return model.MonitorDomain{Domain: strings.TrimSpace(req.Domain), Port: p, Status: "unknown", CheckInterval: i, EnableSsl: req.EnableSSL, VerifyCertificate: req.VerifyCertificate, EnableAlert: req.EnableAlert, ResponseThreshold: req.ResponseThreshold, SslExpiryDays: d}
}

type normalizedMonitorDomain struct {
	domain    string
	port      int64
	enableSSL bool
}

func normalizeMonitorDomainRequest(req *api.MonitorDomainRequest) *api.MonitorDomainRequest {
	normalized := *req
	address := normalizeMonitorDomainAddress(req.Domain, req.Port, req.EnableSSL)
	normalized.Domain, normalized.Port, normalized.EnableSSL = address.domain, address.port, address.enableSSL
	if !normalized.EnableSSL {
		normalized.VerifyCertificate = false
	}
	return &normalized
}

func normalizeMonitorDomainAddress(raw string, port int64, enableSSL bool) normalizedMonitorDomain {
	value := strings.TrimSpace(raw)
	if !strings.Contains(value, "://") {
		return normalizedMonitorDomain{domain: strings.TrimRight(strings.SplitN(value, "/", 2)[0], "/"), port: port, enableSSL: enableSSL}
	}
	parsed, err := url.Parse(value)
	if err != nil || parsed.Hostname() == "" {
		return normalizedMonitorDomain{domain: value, port: port, enableSSL: enableSSL}
	}
	scheme := strings.ToLower(parsed.Scheme)
	if scheme == "https" {
		enableSSL = true
		port = 443
	} else if scheme == "http" {
		enableSSL = false
		port = 80
	}
	if parsed.Port() != "" {
		if parsedPort, parseErr := strconv.ParseInt(parsed.Port(), 10, 64); parseErr == nil {
			port = parsedPort
		}
	}
	return normalizedMonitorDomain{domain: parsed.Hostname(), port: port, enableSSL: enableSSL}
}

func rootCAName(state tls.ConnectionState) string {
	if len(state.VerifiedChains) > 0 && len(state.VerifiedChains[0]) > 0 {
		return certificateName(state.VerifiedChains[0][len(state.VerifiedChains[0])-1].Subject)
	}
	if len(state.PeerCertificates) > 0 {
		return certificateName(state.PeerCertificates[len(state.PeerCertificates)-1].Issuer)
	}
	return ""
}

func certificateName(name pkix.Name) string {
	if name.CommonName != "" {
		return name.CommonName
	}
	if len(name.Organization) > 0 {
		return name.Organization[0]
	}
	return name.String()
}
func monitorDomainResponse(r model.MonitorDomain) *api.MonitorDomainResponse {
	return &api.MonitorDomainResponse{ID: r.ID, Domain: r.Domain, Port: r.Port, Status: r.Status, ResponseTime: r.ResponseTime, SSLValid: r.SslValid, SSLExpiry: r.SslExpiry, SSLRootCA: r.SslRootCa, CheckInterval: r.CheckInterval, EnableSSL: r.EnableSsl, VerifyCertificate: r.VerifyCertificate, EnableAlert: r.EnableAlert, ResponseThreshold: r.ResponseThreshold, SSLExpiryDays: r.SslExpiryDays, LastCheck: r.LastCheck, LastError: r.LastError, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
func monitorError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return err
}

func (b *MonitorBiz) ListChannels(req *api.MonitorAlertChannelListRequest) ([]*api.MonitorAlertChannelResponse, error) {
	var rows []model.MonitorAlertChannel
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*api.MonitorAlertChannelResponse, 0, len(rows))
	for _, r := range rows {
		item, err := b.monitorAlertChannelResponse(r)
		if err != nil {
			return nil, err
		}
		out = append(out, item)
	}
	return out, nil
}
func (b *MonitorBiz) SaveChannel(req *api.MonitorAlertChannelRequest) (*api.MonitorAlertChannelResponse, error) {
	if err := validateAlertChannel(req); err != nil {
		return nil, err
	}
	config := *req.Config
	var err error
	if config.SMTPPassword != "" {
		config.SMTPPassword, err = cryptox.Encrypt(b.aesKey, config.SMTPPassword)
		if err != nil {
			return nil, err
		}
	}
	if config.DingTalkSecret != "" {
		config.DingTalkSecret, err = cryptox.Encrypt(b.aesKey, config.DingTalkSecret)
		if err != nil {
			return nil, err
		}
	}
	encoded, err := json.Marshal(&config)
	if err != nil {
		return nil, err
	}
	row := model.MonitorAlertChannel{ID: req.ID, Name: strings.TrimSpace(req.Name), ChannelType: req.ChannelType, Enabled: req.Enabled, Config: string(encoded)}
	if req.ID > 0 {
		var old model.MonitorAlertChannel
		if err := b.db.First(&old, req.ID).Error; err != nil {
			return nil, monitorError(err)
		}
		row.CreatedAt = old.CreatedAt
	}
	if err := b.db.Save(&row).Error; err != nil {
		return nil, err
	}
	return b.monitorAlertChannelResponse(row)
}

func validateAlertChannel(req *api.MonitorAlertChannelRequest) error {
	if req.Config == nil {
		return errors.New("告警通道配置不能为空")
	}
	c := req.Config
	switch req.ChannelType {
	case "email":
		if c.SMTPHost == "" || c.SMTPPort <= 0 || c.FromEmail == "" || c.SMTPUser == "" {
			return errors.New("请完整填写邮件服务器配置")
		}
	case "webhook":
		if c.WebhookURL == "" {
			return errors.New("请填写 Webhook URL")
		}
	case "wechat":
		if c.WeChatWebhook == "" {
			return errors.New("请填写企业微信 Webhook URL")
		}
	case "dingtalk":
		if c.DingTalkWebhook == "" {
			return errors.New("请填写钉钉 Webhook URL")
		}
	case "feishu":
		if c.FeishuWebhook == "" {
			return errors.New("请填写飞书 Webhook URL")
		}
	default:
		return errors.New("不支持的告警通道类型")
	}
	return nil
}

func (b *MonitorBiz) monitorAlertChannelResponse(row model.MonitorAlertChannel) (*api.MonitorAlertChannelResponse, error) {
	config := new(api.MonitorAlertChannelConfig)
	if row.Config != "" {
		if err := json.Unmarshal([]byte(row.Config), config); err != nil {
			return nil, fmt.Errorf("解析告警通道配置: %w", err)
		}
	}
	var err error
	if config.SMTPPassword != "" {
		config.SMTPPassword, err = cryptox.Decrypt(b.aesKey, config.SMTPPassword)
		if err != nil {
			return nil, err
		}
	}
	if config.DingTalkSecret != "" {
		config.DingTalkSecret, err = cryptox.Decrypt(b.aesKey, config.DingTalkSecret)
		if err != nil {
			return nil, err
		}
	}
	return &api.MonitorAlertChannelResponse{ID: row.ID, Name: row.Name, ChannelType: row.ChannelType, Enabled: row.Enabled, Config: config, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt}, nil
}
func (b *MonitorBiz) DeleteChannel(req *api.MonitorAlertChannelIDRequest) error {
	r := b.db.Delete(&model.MonitorAlertChannel{}, req.ID)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *MonitorBiz) ListReceivers(req *api.MonitorAlertReceiverListRequest) ([]*api.MonitorAlertReceiverResponse, error) {
	var rows []model.MonitorAlertReceiver
	q := b.db.Order("id DESC")
	if req.Keyword != "" {
		q = q.Where("name LIKE ? OR email LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*api.MonitorAlertReceiverResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, monitorAlertReceiverResponse(r))
	}
	return out, nil
}
func (b *MonitorBiz) SaveReceiver(req *api.MonitorAlertReceiverRequest) (*api.MonitorAlertReceiverResponse, error) {
	if err := validateAlertReceiver(req); err != nil {
		return nil, err
	}
	row := model.MonitorAlertReceiver{ID: req.ID, Name: strings.TrimSpace(req.Name), Email: strings.TrimSpace(req.Email), Phone: strings.TrimSpace(req.Phone), Webhook: strings.TrimSpace(req.Webhook), WechatID: strings.TrimSpace(req.WechatID), DingtalkID: strings.TrimSpace(req.DingtalkID), FeishuID: strings.TrimSpace(req.FeishuID), EnableEmail: req.EnableEmail, EnableWebhook: req.EnableWebhook, EnableWechat: req.EnableWechat, EnableDingtalk: req.EnableDingtalk, EnableFeishu: req.EnableFeishu, EnableSystemMsg: req.EnableSystemMsg, Enabled: req.Enabled}
	if req.ID > 0 {
		var old model.MonitorAlertReceiver
		if err := b.db.First(&old, req.ID).Error; err != nil {
			return nil, monitorError(err)
		}
		row.CreatedAt = old.CreatedAt
	}
	if err := b.db.Save(&row).Error; err != nil {
		return nil, err
	}
	return monitorAlertReceiverResponse(row), nil
}

func validateAlertReceiver(req *api.MonitorAlertReceiverRequest) error {
	if !req.EnableEmail && !req.EnableWebhook && !req.EnableWechat && !req.EnableDingtalk && !req.EnableFeishu && !req.EnableSystemMsg {
		return errors.New("请至少选择一种通知方式")
	}
	if req.EnableEmail && strings.TrimSpace(req.Email) == "" {
		return errors.New("启用邮件通知时必须填写邮箱地址")
	}
	if req.EnableWebhook && strings.TrimSpace(req.Webhook) == "" {
		return errors.New("启用 Webhook 通知时必须填写 Webhook 地址")
	}
	if req.EnableWechat && strings.TrimSpace(req.WechatID) == "" {
		return errors.New("启用企业微信通知时必须填写企业微信 ID")
	}
	if req.EnableDingtalk && strings.TrimSpace(req.DingtalkID) == "" && strings.TrimSpace(req.Phone) == "" {
		return errors.New("启用钉钉通知时必须填写钉钉 ID 或手机号码")
	}
	if req.EnableFeishu && strings.TrimSpace(req.FeishuID) == "" {
		return errors.New("启用飞书通知时必须填写飞书 ID")
	}
	return nil
}

func monitorAlertReceiverResponse(r model.MonitorAlertReceiver) *api.MonitorAlertReceiverResponse {
	return &api.MonitorAlertReceiverResponse{ID: r.ID, Name: r.Name, Email: r.Email, Phone: r.Phone, Webhook: r.Webhook, WechatID: r.WechatID, DingtalkID: r.DingtalkID, FeishuID: r.FeishuID, EnableEmail: r.EnableEmail, EnableWebhook: r.EnableWebhook, EnableWechat: r.EnableWechat, EnableDingtalk: r.EnableDingtalk, EnableFeishu: r.EnableFeishu, EnableSystemMsg: r.EnableSystemMsg, Enabled: r.Enabled, CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt}
}
func (b *MonitorBiz) DeleteReceiver(req *api.MonitorAlertReceiverIDRequest) error {
	r := b.db.Delete(&model.MonitorAlertReceiver{}, req.ID)
	if r.Error != nil {
		return r.Error
	}
	if r.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *MonitorBiz) ListLogs(req *api.MonitorAlertLogListRequest) ([]*api.MonitorAlertLogResponse, error) {
	var rows []model.MonitorAlertLog
	q := b.db.Order("sent_at DESC")
	if req.DomainID > 0 {
		q = q.Where("domain_id = ?", req.DomainID)
	}
	if req.Status != "" {
		q = q.Where("status = ?", req.Status)
	}
	if err := q.Find(&rows).Error; err != nil {
		return nil, err
	}
	out := make([]*api.MonitorAlertLogResponse, 0, len(rows))
	for _, r := range rows {
		out = append(out, &api.MonitorAlertLogResponse{ID: r.ID, DomainID: r.DomainID, AlertType: r.AlertType, Status: r.Status, Message: r.Message, ChannelType: r.ChannelType, ErrorMessage: r.ErrorMessage, SentAt: r.SentAt})
	}
	return out, nil
}
