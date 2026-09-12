package biz

import (
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gorm.io/gorm"
)

type AuditBiz struct{ db *gorm.DB }

func NewAuditBiz(db *gorm.DB) *AuditBiz { return &AuditBiz{db} }
func (b *AuditBiz) Operations(r *api.AuditListRequest) ([]*api.AuditOperationLogResponse, error) {
	var a []model.AuditOperationLog
	q := b.db.Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("username LIKE ? OR module LIKE ?", "%"+r.Keyword+"%", "%"+r.Keyword+"%")
	}
	if e := q.Find(&a).Error; e != nil {
		return nil, e
	}
	o := make([]*api.AuditOperationLogResponse, 0, len(a))
	for _, x := range a {
		o = append(o, &api.AuditOperationLogResponse{ID: x.ID, Username: x.Username, Module: x.Module, Action: x.Action, Description: x.Description, Method: x.Method, Path: x.Path, Status: x.Status, CostTime: x.CostTime, IP: x.IP, CreatedAt: x.CreatedAt})
	}
	return o, nil
}
func (b *AuditBiz) Logins(r *api.AuditListRequest) ([]*api.AuditLoginLogResponse, error) {
	var a []model.AuditLoginLog
	q := b.db.Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("username LIKE ?", "%"+r.Keyword+"%")
	}
	if e := q.Find(&a).Error; e != nil {
		return nil, e
	}
	o := make([]*api.AuditLoginLogResponse, 0, len(a))
	for _, x := range a {
		o = append(o, &api.AuditLoginLogResponse{ID: x.ID, Username: x.Username, LoginType: x.LoginType, Status: x.Status, IP: x.IP, Message: x.Message, CreatedAt: x.CreatedAt})
	}
	return o, nil
}
func (b *AuditBiz) Data(r *api.AuditListRequest) ([]*api.AuditDataLogResponse, error) {
	var a []model.AuditDataLog
	q := b.db.Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("username LIKE ? OR table_name LIKE ?", "%"+r.Keyword+"%", "%"+r.Keyword+"%")
	}
	if e := q.Find(&a).Error; e != nil {
		return nil, e
	}
	o := make([]*api.AuditDataLogResponse, 0, len(a))
	for _, x := range a {
		o = append(o, &api.AuditDataLogResponse{ID: x.ID, Username: x.Username, TableName: x.TableName_, RecordID: x.RecordID, Action: x.Action, OldData: x.OldData, NewData: x.NewData, CreatedAt: x.CreatedAt})
	}
	return o, nil
}
func (b *AuditBiz) DeleteOperations(r *api.AuditDeleteRequest) error {
	return b.db.Where("id IN ?", r.IDs).Delete(&model.AuditOperationLog{}).Error
}
func (b *AuditBiz) DeleteLogins(r *api.AuditDeleteRequest) error {
	return b.db.Where("id IN ?", r.IDs).Delete(&model.AuditLoginLog{}).Error
}
func (b *AuditBiz) DeleteData(r *api.AuditDeleteRequest) error {
	return b.db.Where("id IN ?", r.IDs).Delete(&model.AuditDataLog{}).Error
}
