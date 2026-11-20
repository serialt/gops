package biz

import (
	"context"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/repo"

	"gorm.io/gorm"
)

type DashboardBiz struct {
	db   *gorm.DB
	repo *repo.DashboardRepo
}

func NewDashboardBiz(db *gorm.DB, dashboardRepo *repo.DashboardRepo) *DashboardBiz {
	return &DashboardBiz{db: db, repo: dashboardRepo}
}

func (b *DashboardBiz) Stats(ctx context.Context) (*api.DashboardStatsResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := b.repo.Stats(ctx, b.db, todayStart)
	if err != nil {
		return nil, err
	}
	certificates := make([]*api.DashboardDomainCertificateResponse, 0, len(stats.DomainCertificates))
	for _, certificate := range stats.DomainCertificates {
		certificates = append(certificates, &api.DashboardDomainCertificateResponse{Domain: certificate.Domain, SSLExpiry: certificate.SSLExpiry})
	}
	return &api.DashboardStatsResponse{
		HostsTotal: stats.HostsTotal, HostsOnline: stats.HostsOnline, HostsOffline: stats.HostsTotal - stats.HostsOnline,
		ClustersTotal: stats.ClustersTotal, ClustersOffline: stats.ClustersOffline,
		NodesTotal: stats.NodesTotal, NodesOffline: stats.NodesOffline,
		PodsTotal: stats.PodsTotal, PodsAbnormal: stats.PodsAbnormal,
		DomainsTotal: stats.DomainsTotal, DomainsNormal: stats.DomainsNormal,
		DomainsAbnormal: stats.DomainsAbnormal, DomainsPaused: stats.DomainsPaused,
		DomainCertificates: certificates,
		TodayOperations:    stats.TodayOperations, ActiveAlerts: stats.ActiveAlerts,
	}, nil
}
