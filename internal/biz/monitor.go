package biz

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gorm.io/gorm"
)

type MonitorBiz struct {
	db     *gorm.DB
	client *http.Client
}

func NewMonitorBiz(db *gorm.DB) *MonitorBiz {
	return &MonitorBiz{db: db, client: &http.Client{Timeout: 15 * time.Second}}
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
	row := b.domainModel(req)
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
	updated := b.domainModel(req)
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
	start := time.Now()
	scheme := "http"
	if row.EnableSsl {
		scheme = "https"
	}
	res, err := b.client.Get(fmt.Sprintf("%s://%s:%d", scheme, row.Domain, row.Port))
	row.LastCheck, row.ResponseTime, row.NextCheck = time.Now(), time.Since(start).Milliseconds(), time.Now().Add(time.Duration(row.CheckInterval)*time.Second)
	row.LastError = ""
	row.Status = "normal"
	if err != nil {
		row.Status, row.LastError = "abnormal", err.Error()
	} else {
		_ = res.Body.Close()
		if res.StatusCode >= 400 {
			row.Status, row.LastError = "abnormal", res.Status
		}
	}
	if err := b.db.Save(&row).Error; err != nil {
		return nil, err
	}
	if row.Status == "abnormal" && row.EnableAlert {
		_ = b.db.Create(&model.MonitorAlertLog{DomainID: row.ID, AlertType: "availability", Status: "pending", Message: row.LastError, SentAt: time.Now()}).Error
	}
	return monitorDomainResponse(row), nil
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
	return model.MonitorDomain{Domain: strings.TrimSpace(req.Domain), Port: p, Status: "unknown", CheckInterval: i, EnableSsl: req.EnableSSL, EnableAlert: req.EnableAlert, ResponseThreshold: req.ResponseThreshold, SslExpiryDays: d}
}
func monitorDomainResponse(r model.MonitorDomain) *api.MonitorDomainResponse {
	return &api.MonitorDomainResponse{ID: r.ID, Domain: r.Domain, Port: r.Port, Status: r.Status, ResponseTime: r.ResponseTime, SSLValid: r.SslValid, SSLExpiry: r.SslExpiry, CheckInterval: r.CheckInterval, EnableSSL: r.EnableSsl, EnableAlert: r.EnableAlert, ResponseThreshold: r.ResponseThreshold, SSLExpiryDays: r.SslExpiryDays, LastCheck: r.LastCheck, LastError: r.LastError, UpdatedAt: r.UpdatedAt}
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
		out = append(out, &api.MonitorAlertChannelResponse{ID: r.ID, Name: r.Name, ChannelType: r.ChannelType, Enabled: r.Enabled, Config: r.Config, UpdatedAt: r.UpdatedAt})
	}
	return out, nil
}
func (b *MonitorBiz) SaveChannel(req *api.MonitorAlertChannelRequest) (*api.MonitorAlertChannelResponse, error) {
	row := model.MonitorAlertChannel{ID: req.ID, Name: req.Name, ChannelType: req.ChannelType, Enabled: req.Enabled, Config: req.Config}
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
	return &api.MonitorAlertChannelResponse{ID: row.ID, Name: row.Name, ChannelType: row.ChannelType, Enabled: row.Enabled, Config: row.Config, UpdatedAt: row.UpdatedAt}, nil
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
		out = append(out, &api.MonitorAlertReceiverResponse{ID: r.ID, Name: r.Name, Email: r.Email, Phone: r.Phone, Webhook: r.Webhook, Enabled: r.Enabled, UpdatedAt: r.UpdatedAt})
	}
	return out, nil
}
func (b *MonitorBiz) SaveReceiver(req *api.MonitorAlertReceiverRequest) (*api.MonitorAlertReceiverResponse, error) {
	row := model.MonitorAlertReceiver{ID: req.ID, Name: req.Name, Email: req.Email, Phone: req.Phone, Webhook: req.Webhook, Enabled: req.Enabled}
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
	return &api.MonitorAlertReceiverResponse{ID: row.ID, Name: row.Name, Email: row.Email, Phone: row.Phone, Webhook: row.Webhook, Enabled: row.Enabled, UpdatedAt: row.UpdatedAt}, nil
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
