package repo

import (
	"context"
	"time"

	"gops/internal/data/query"

	"gorm.io/gorm"
)

type DashboardStats struct {
	HostsTotal         int64
	HostsOnline        int64
	ClustersTotal      int64
	ClustersOffline    int64
	NodesTotal         int64
	NodesOffline       int64
	PodsTotal          int64
	PodsAbnormal       int64
	DomainsTotal       int64
	DomainsNormal      int64
	DomainsAbnormal    int64
	DomainsPaused      int64
	DomainCertificates []*DashboardDomainCertificate
	TodayOperations    int64
	ActiveAlerts       int64
}

type DashboardDomainCertificate struct {
	Domain    string
	SSLExpiry time.Time
}

type DashboardRepo struct{}

func NewDashboardRepo() *DashboardRepo { return &DashboardRepo{} }

func (r *DashboardRepo) Stats(ctx context.Context, db *gorm.DB, todayStart time.Time) (*DashboardStats, error) {
	q := query.Use(db)
	result := new(DashboardStats)
	var err error
	if result.HostsTotal, err = q.Host.WithContext(ctx).Count(); err != nil {
		return nil, err
	}
	if result.HostsOnline, err = q.Host.WithContext(ctx).Where(q.Host.Status.Eq(1)).Count(); err != nil {
		return nil, err
	}
	clusters, err := q.KubernetesCluster.WithContext(ctx).Find()
	if err != nil {
		return nil, err
	}
	result.ClustersTotal = int64(len(clusters))
	for _, cluster := range clusters {
		if cluster.Status != 1 {
			result.ClustersOffline++
			result.NodesOffline += cluster.NodeCount
			result.PodsAbnormal += cluster.PodCount
		}
		result.NodesTotal += cluster.NodeCount
		result.PodsTotal += cluster.PodCount
	}
	if result.DomainsTotal, err = q.MonitorDomain.WithContext(ctx).Count(); err != nil {
		return nil, err
	}
	if result.DomainsNormal, err = q.MonitorDomain.WithContext(ctx).Where(q.MonitorDomain.Status.Eq("normal")).Count(); err != nil {
		return nil, err
	}
	if result.DomainsAbnormal, err = q.MonitorDomain.WithContext(ctx).Where(q.MonitorDomain.Status.Eq("abnormal")).Count(); err != nil {
		return nil, err
	}
	if result.DomainsPaused, err = q.MonitorDomain.WithContext(ctx).Where(q.MonitorDomain.Status.Eq("paused")).Count(); err != nil {
		return nil, err
	}
	certificateRows, err := q.MonitorDomain.WithContext(ctx).
		Where(q.MonitorDomain.EnableSsl.Is(true), q.MonitorDomain.SslExpiry.Gt(time.Time{})).
		Order(q.MonitorDomain.SslExpiry.Asc()).
		Limit(10).
		Find()
	if err != nil {
		return nil, err
	}
	result.DomainCertificates = make([]*DashboardDomainCertificate, 0, len(certificateRows))
	for _, row := range certificateRows {
		result.DomainCertificates = append(result.DomainCertificates, &DashboardDomainCertificate{Domain: row.Domain, SSLExpiry: row.SslExpiry})
	}
	if result.TodayOperations, err = q.AuditOperationLog.WithContext(ctx).Where(q.AuditOperationLog.CreatedAt.Gte(todayStart)).Count(); err != nil {
		return nil, err
	}
	if result.ActiveAlerts, err = q.MonitorAlertLog.WithContext(ctx).Where(q.MonitorAlertLog.Status.Eq("pending")).Count(); err != nil {
		return nil, err
	}
	return result, nil
}
