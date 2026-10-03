package biz

import (
	"strings"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/model"

	"gorm.io/gorm"
)

type AuditBiz struct{ db *gorm.DB }

func NewAuditBiz(db *gorm.DB) *AuditBiz { return &AuditBiz{db} }
func (b *AuditBiz) Operations(r *api.AuditListRequest) (*api.PageResponse[*api.AuditOperationLogResponse], error) {
	q := b.db.Model(&model.AuditOperationLog{}).Order("id DESC")
	if r.Keyword != "" {
		like := "%" + strings.TrimSpace(r.Keyword) + "%"
		q = q.Where("username LIKE ? OR module LIKE ? OR action LIKE ? OR description LIKE ? OR path LIKE ? OR ip LIKE ?", like, like, like, like, like, like)
	}
	if r.Username != "" {
		q = q.Where("username LIKE ?", "%"+strings.TrimSpace(r.Username)+"%")
	}
	if r.Module != "" {
		q = q.Where("module = ?", r.Module)
	}
	if r.Action != "" {
		q = q.Where("action = ?", r.Action)
	}
	if len(r.Status) == 3 && strings.HasSuffix(r.Status, "xx") && r.Status[0] >= '1' && r.Status[0] <= '5' {
		lower := int64(r.Status[0]-'0') * 100
		q = q.Where("status >= ? AND status < ?", lower, lower+100)
	}
	q = applyAuditTimeRange(q, r.StartTime, r.EndTime)
	a, total, e := queryPage[model.AuditOperationLog](q, r.Page, r.Limit)
	if e != nil {
		return nil, e
	}
	o := make([]*api.AuditOperationLogResponse, 0, len(a))
	for _, x := range a {
		o = append(o, &api.AuditOperationLogResponse{ID: x.ID, Username: x.Username, Module: x.Module, Action: x.Action, Description: x.Description, Method: x.Method, Path: x.Path, Status: x.Status, CostTime: x.CostTime, IP: x.IP, CreatedAt: x.CreatedAt})
	}
	return &api.PageResponse[*api.AuditOperationLogResponse]{List: o, Total: total}, nil
}
func (b *AuditBiz) Logins(r *api.AuditListRequest) (*api.PageResponse[*api.AuditLoginLogResponse], error) {
	q := b.db.Model(&model.AuditLoginLog{}).Order("id DESC")
	if r.Keyword != "" {
		like := "%" + strings.TrimSpace(r.Keyword) + "%"
		q = q.Where("username LIKE ? OR ip LIKE ? OR message LIKE ?", like, like, like)
	}
	if r.Username != "" {
		q = q.Where("username LIKE ?", "%"+strings.TrimSpace(r.Username)+"%")
	}
	if r.LoginType != "" {
		q = q.Where("login_type = ?", r.LoginType)
	}
	if r.Status != "" {
		q = q.Where("status = ?", r.Status)
	}
	q = applyAuditTimeRange(q, r.StartTime, r.EndTime)
	a, total, e := queryPage[model.AuditLoginLog](q, r.Page, r.Limit)
	if e != nil {
		return nil, e
	}
	o := make([]*api.AuditLoginLogResponse, 0, len(a))
	for _, x := range a {
		o = append(o, &api.AuditLoginLogResponse{ID: x.ID, Username: x.Username, LoginType: x.LoginType, Status: x.Status, IP: x.IP, Message: x.Message, CreatedAt: x.CreatedAt})
	}
	return &api.PageResponse[*api.AuditLoginLogResponse]{List: o, Total: total}, nil
}
func (b *AuditBiz) DeleteOperations(r *api.AuditDeleteRequest) error {
	return b.db.Where("id IN ?", r.IDs).Delete(&model.AuditOperationLog{}).Error
}
func (b *AuditBiz) DeleteLogins(r *api.AuditDeleteRequest) error {
	return b.db.Where("id IN ?", r.IDs).Delete(&model.AuditLoginLog{}).Error
}
func applyAuditTimeRange(q *gorm.DB, start, end string) *gorm.DB {
	if startTime, err := time.Parse("2006-01-02", start); err == nil {
		q = q.Where("created_at >= ?", startTime)
	}
	if endTime, err := time.Parse("2006-01-02", end); err == nil {
		q = q.Where("created_at < ?", endTime.AddDate(0, 0, 1))
	}
	return q
}
