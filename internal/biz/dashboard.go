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
	return &api.DashboardStatsResponse{
		HostsTotal: stats.HostsTotal, HostsOnline: stats.HostsOnline, HostsOffline: stats.HostsTotal - stats.HostsOnline,
		ClustersTotal: stats.ClustersTotal, NodesTotal: stats.NodesTotal, PodsTotal: stats.PodsTotal,
		TodayOperations: stats.TodayOperations, ActiveAlerts: stats.ActiveAlerts,
	}, nil
}
