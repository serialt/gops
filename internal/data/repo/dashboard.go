package repo

import (
	"context"
	"time"

	"gops/internal/data/query"

	"gorm.io/gorm"
)

type DashboardStats struct {
	HostsTotal      int64
	HostsOnline     int64
	ClustersTotal   int64
	NodesTotal      int64
	PodsTotal       int64
	TodayOperations int64
	ActiveAlerts    int64
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
		result.NodesTotal += cluster.NodeCount
		result.PodsTotal += cluster.PodCount
	}
	if result.TodayOperations, err = q.AuditOperationLog.WithContext(ctx).Where(q.AuditOperationLog.CreatedAt.Gte(todayStart)).Count(); err != nil {
		return nil, err
	}
	if result.ActiveAlerts, err = q.MonitorAlertLog.WithContext(ctx).Where(q.MonitorAlertLog.Status.Eq("pending")).Count(); err != nil {
		return nil, err
	}
	return result, nil
}
