package repo

import (
	"context"
	"time"

	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gorm"
)

type MetricsRepo struct{}

func NewMetricsRepo() *MetricsRepo { return &MetricsRepo{} }

func (r *MetricsRepo) RecentLogins(ctx context.Context, db *gorm.DB, limit int) ([]*model.AuditLoginLog, error) {
	q := query.Use(db).AuditLoginLog
	return q.WithContext(ctx).Order(q.CreatedAt.Desc()).Limit(limit).Find()
}

func (r *MetricsRepo) LoginLogsSince(ctx context.Context, db *gorm.DB, start time.Time) ([]*model.AuditLoginLog, error) {
	q := query.Use(db).AuditLoginLog
	return q.WithContext(ctx).Where(q.CreatedAt.Gte(start)).Order(q.CreatedAt.Desc(), q.ID.Desc()).Limit(10000).Find()
}

func (r *MetricsRepo) OperationLogsSince(ctx context.Context, db *gorm.DB, start time.Time) ([]*model.AuditOperationLog, error) {
	q := query.Use(db).AuditOperationLog
	return q.WithContext(ctx).Where(q.CreatedAt.Gte(start)).Order(q.ID.Desc()).Limit(10000).Find()
}

func (r *MetricsRepo) Hosts(ctx context.Context, db *gorm.DB) ([]*model.Host, error) {
	q := query.Use(db).Host
	return q.WithContext(ctx).Order(q.ID.Asc()).Find()
}

func (r *MetricsRepo) TerminalSessions(ctx context.Context, db *gorm.DB) ([]*model.TerminalSession, error) {
	q := query.Use(db).TerminalSession
	return q.WithContext(ctx).Order(q.StartedAt.Desc(), q.ID.Desc()).Limit(10000).Find()
}

func (r *MetricsRepo) Users(ctx context.Context, db *gorm.DB, ids []int64) ([]*model.User, error) {
	if len(ids) == 0 {
		return []*model.User{}, nil
	}
	q := query.Use(db).User
	return q.WithContext(ctx).Where(q.ID.In(ids...)).Find()
}

func (r *MetricsRepo) KubernetesClusters(ctx context.Context, db *gorm.DB) ([]*model.KubernetesCluster, error) {
	q := query.Use(db).KubernetesCluster
	return q.WithContext(ctx).Order(q.ID.Asc()).Find()
}

func (r *MetricsRepo) CountMonitorDomains(ctx context.Context, db *gorm.DB) (int64, error) {
	q := query.Use(db).MonitorDomain
	return q.WithContext(ctx).Count()
}

func (r *MetricsRepo) CountValidMonitorCertificates(ctx context.Context, db *gorm.DB) (int64, error) {
	q := query.Use(db).MonitorDomain
	return q.WithContext(ctx).Where(q.EnableSsl.Is(true), q.SslValid.Is(true)).Count()
}

func (r *MetricsRepo) CountExpiringSSLCertificates(ctx context.Context, db *gorm.DB, now, deadline time.Time) (int64, error) {
	q := query.Use(db).SslCertificate
	return q.WithContext(ctx).
		Where(q.Certificate.Neq(""), q.NotAfter.Gt(now), q.NotAfter.Lte(deadline)).
		Count()
}

func (r *MetricsRepo) CountExpiringMonitorCertificates(ctx context.Context, db *gorm.DB, now, deadline time.Time) (int64, error) {
	q := query.Use(db).MonitorDomain
	return q.WithContext(ctx).
		Where(q.EnableSsl.Is(true), q.SslValid.Is(true), q.SslExpiry.Gt(now), q.SslExpiry.Lte(deadline)).
		Count()
}
