package biz

import (
	"gops/internal/data/api"
	"gops/internal/data/model"
	"gorm.io/gorm"
)

type SSLCertBiz struct{ db *gorm.DB }

func NewSSLCertBiz(db *gorm.DB) *SSLCertBiz { return &SSLCertBiz{db} }
func (b *SSLCertBiz) List(r *api.SSLCertificateListRequest) (*api.PageResponse[*api.SSLCertificateResponse], error) {
	q := b.db.Model(&model.SslCertificate{}).Order("id DESC")
	if r.Keyword != "" {
		q = q.Where("name LIKE ? OR domain LIKE ?", "%"+r.Keyword+"%", "%"+r.Keyword+"%")
	}
	a, total, e := queryPage[model.SslCertificate](q, r.Page, r.Limit)
	if e != nil {
		return nil, e
	}
	o := make([]*api.SSLCertificateResponse, 0, len(a))
	for _, x := range a {
		o = append(o, sslCert(x))
	}
	return &api.PageResponse[*api.SSLCertificateResponse]{List: o, Total: total}, nil
}
func (b *SSLCertBiz) Save(r *api.SSLCertificateRequest) (*api.SSLCertificateResponse, error) {
	x := model.SslCertificate{ID: r.ID, Name: r.Name, Domain: r.Domain, SourceType: r.SourceType, CaProvider: r.CaProvider, AutoRenew: r.AutoRenew, RenewDaysBefore: r.RenewDaysBefore, Status: "pending"}
	if r.ID > 0 {
		var old model.SslCertificate
		if e := b.db.First(&old, r.ID).Error; e != nil {
			return nil, e
		}
		x.CreatedAt = old.CreatedAt
		x.Status = old.Status
	}
	if e := b.db.Save(&x).Error; e != nil {
		return nil, e
	}
	return sslCert(x), nil
}
func sslCert(x model.SslCertificate) *api.SSLCertificateResponse {
	return &api.SSLCertificateResponse{ID: x.ID, Name: x.Name, Domain: x.Domain, SourceType: x.SourceType, CaProvider: x.CaProvider, Status: x.Status, AutoRenew: x.AutoRenew, RenewDaysBefore: x.RenewDaysBefore, NotAfter: x.NotAfter}
}
