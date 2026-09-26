package biz

import (
	"errors"
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gorm.io/gorm"
)

type NginxBiz struct{ db *gorm.DB }

func NewNginxBiz(db *gorm.DB) *NginxBiz { return &NginxBiz{db} }
func (b *NginxBiz) ListSources(r *api.NginxSourceListRequest) (*api.PageResponse[*api.NginxSourceResponse], error) {
	q := b.db.Model(&model.NginxSource{}).Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("name LIKE ?", "%"+r.Keyword+"%")
	}
	rows, total, e := queryPage[model.NginxSource](q, r.Page, r.Limit)
	if e != nil {
		return nil, e
	}
	out := make([]*api.NginxSourceResponse, 0, len(rows))
	for _, x := range rows {
		out = append(out, nginxSource(x))
	}
	return &api.PageResponse[*api.NginxSourceResponse]{List: out, Total: total}, nil
}
func (b *NginxBiz) SaveSource(r *api.NginxSourceRequest) (*api.NginxSourceResponse, error) {
	x := model.NginxSource{ID: r.ID, Name: r.Name, SourceType: r.SourceType, Description: r.Description, Status: r.Status, HostID: r.HostID, LogPath: r.LogPath, LogFormat: r.LogFormat, CollectInterval: r.CollectInterval, RetentionDays: r.RetentionDays}
	if x.SourceType == "" {
		x.SourceType = "host"
	}
	if x.LogFormat == "" {
		x.LogFormat = "combined"
	}
	if x.CollectInterval == 0 {
		x.CollectInterval = 60
	}
	if x.RetentionDays == 0 {
		x.RetentionDays = 30
	}
	if r.ID > 0 {
		var old model.NginxSource
		if e := b.db.First(&old, r.ID).Error; e != nil {
			return nil, nginxErr(e)
		}
		x.CreatedAt = old.CreatedAt
	}
	if e := b.db.Save(&x).Error; e != nil {
		return nil, e
	}
	return nginxSource(x), nil
}
func (b *NginxBiz) DeleteSource(r *api.NginxSourceIDRequest) error {
	x := b.db.Delete(&model.NginxSource{}, r.ID)
	if x.Error != nil {
		return x.Error
	}
	if x.RowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}
func (b *NginxBiz) Overview() (*api.NginxOverviewResponse, error) {
	o := &api.NginxOverviewResponse{}
	for _, q := range []struct {
		m any
		w string
		p *int64
	}{{&model.NginxSource{}, "", &o.Sources}, {&model.NginxSource{}, "status = 1", &o.ActiveSources}, {&model.NginxAccessLog{}, "", &o.Requests}, {&model.NginxAccessLog{}, "status >= 400", &o.ErrorRequests}} {
		d := b.db.Model(q.m)
		if q.w != "" {
			d = d.Where(q.w)
		}
		if e := d.Count(q.p).Error; e != nil {
			return nil, e
		}
	}
	return o, nil
}
func (b *NginxBiz) ListLogs(r *api.NginxAccessLogListRequest) (*api.PageResponse[*api.NginxAccessLogResponse], error) {
	q := b.db.Model(&model.NginxAccessLog{}).Order("occurred_at DESC")
	if r.SourceID > 0 {
		q = q.Where("source_id = ?", r.SourceID)
	}
	if r.Keyword != "" {
		q = q.Where("uri LIKE ? OR remote_addr LIKE ?", "%"+r.Keyword+"%", "%"+r.Keyword+"%")
	}
	rows, total, e := queryPage[model.NginxAccessLog](q, r.Page, r.Limit)
	if e != nil {
		return nil, e
	}
	out := make([]*api.NginxAccessLogResponse, 0, len(rows))
	for _, x := range rows {
		out = append(out, &api.NginxAccessLogResponse{ID: x.ID, SourceID: x.SourceID, OccurredAt: x.OccurredAt, RemoteAddr: x.RemoteAddr, Method: x.Method, URI: x.URI, Status: x.Status, BodyBytesSent: x.BodyBytesSent, RequestTime: x.RequestTime, UserAgent: x.UserAgent})
	}
	return &api.PageResponse[*api.NginxAccessLogResponse]{List: out, Total: total}, nil
}
func nginxSource(x model.NginxSource) *api.NginxSourceResponse {
	return &api.NginxSourceResponse{ID: x.ID, Name: x.Name, SourceType: x.SourceType, Description: x.Description, Status: x.Status, HostID: x.HostID, LogPath: x.LogPath, LogFormat: x.LogFormat, CollectInterval: x.CollectInterval, RetentionDays: x.RetentionDays, LastCollectLogs: x.LastCollectLogs, LastError: x.LastError, UpdatedAt: x.UpdatedAt}
}
func nginxErr(e error) error {
	if errors.Is(e, gorm.ErrRecordNotFound) {
		return ErrNotFound
	}
	return e
}
