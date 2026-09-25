package biz

import (
	"context"
	"time"

	"gops/internal/data/api"
	"gops/internal/data/repo"
)

type DashboardBiz struct {
	repo *repo.DashboardRepo
}

func NewDashboardBiz(dashboardRepo *repo.DashboardRepo) *DashboardBiz {
	return &DashboardBiz{repo: dashboardRepo}
}

func (b *DashboardBiz) Stats(ctx context.Context) (*api.DashboardStatsResponse, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	stats, err := b.repo.Stats(ctx, todayStart)
	if err != nil {
		return nil, err
	}
	return &api.DashboardStatsResponse{
		HostsTotal: stats.HostsTotal, HostsOnline: stats.HostsOnline, HostsOffline: stats.HostsTotal - stats.HostsOnline,
		ClustersTotal: stats.ClustersTotal, NodesTotal: stats.NodesTotal, PodsTotal: stats.PodsTotal,
		TodayOperations: stats.TodayOperations, ActiveAlerts: stats.ActiveAlerts,
	}, nil
}
