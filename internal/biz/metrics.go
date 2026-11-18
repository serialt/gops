package biz

import (
	"context"
	"fmt"
	"sort"
	"time"

	"gops/internal/data/repo"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"gorm.io/gorm"
)

type MetricsRecentLogin struct {
	ID        int64
	Username  string
	LoginType string
	IP        string
	Status    string
	CreatedAt time.Time
}

type MetricsLoginTypeCount struct {
	LoginType string
	Status    string
	Count     int64
}

type MetricsHost struct {
	ID                int64
	Name              string
	IP                string
	OS                string
	Hostname          string
	Status            int64
	CPUUsage          float64
	MemoryUsage       float64
	DiskUsage         float64
	RecentUser        string
	RecentSessionType string
	RecentStatus      string
	RecentAt          time.Time
	HasRecent         bool
}

type MetricsKubernetesCluster struct {
	ID        int64
	Name      string
	Alias     string
	Status    int64
	NodeCount int64
	PodCount  int64
}

type MetricsSnapshot struct {
	RecentLogins             []MetricsRecentLogin
	FailedUsersDaily         int64
	SuccessUsersDaily        int64
	LoginTypeCounts          []MetricsLoginTypeCount
	OperationStatusDaily     map[int64]int64
	MonitorDomainsTotal      int64
	MonitorCertificatesValid int64
	CertificateExpiring      map[string]map[string]int64
	Hosts                    []MetricsHost
	KubernetesClusters       []MetricsKubernetesCluster
	KubernetesHealthy        int64
	KubernetesNodesTotal     int64
}

type MetricsBiz struct {
	db   *gorm.DB
	repo *repo.MetricsRepo
}

func NewMetricsBiz(db *gorm.DB, repo *repo.MetricsRepo) *MetricsBiz {
	return &MetricsBiz{db: db, repo: repo}
}

func (b *MetricsBiz) NewPrometheusRegistry(businessCollector prometheus.Collector) *prometheus.Registry {
	registry := prometheus.NewRegistry()
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	if businessCollector != nil {
		registry.MustRegister(businessCollector)
	}
	return registry
}

func (b *MetricsBiz) Snapshot(ctx context.Context) (*MetricsSnapshot, error) {
	now := time.Now()
	start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	result := &MetricsSnapshot{OperationStatusDaily: make(map[int64]int64), CertificateExpiring: make(map[string]map[string]int64)}

	recent, err := b.repo.RecentLogins(ctx, b.db, 100)
	if err != nil {
		return nil, err
	}
	for _, row := range recent {
		result.RecentLogins = append(result.RecentLogins, MetricsRecentLogin{ID: row.ID, Username: row.Username, LoginType: row.LoginType, IP: row.IP, Status: row.Status, CreatedAt: row.CreatedAt})
	}

	loginRows, err := b.repo.LoginLogsSince(ctx, b.db, start)
	if err != nil {
		return nil, err
	}
	failedUsers := make(map[string]struct{})
	successUsers := make(map[string]struct{})
	typeUsers := make(map[string]map[string]map[string]struct{})
	for _, row := range loginRows {
		if row.Username == "" {
			continue
		}
		if row.Status == "failed" {
			failedUsers[row.Username] = struct{}{}
		}
		if row.Status == "success" {
			successUsers[row.Username] = struct{}{}
		}
		loginType := row.LoginType
		if loginType == "" {
			loginType = "unknown"
		}
		if typeUsers[loginType] == nil {
			typeUsers[loginType] = make(map[string]map[string]struct{})
		}
		if typeUsers[loginType][row.Status] == nil {
			typeUsers[loginType][row.Status] = make(map[string]struct{})
		}
		typeUsers[loginType][row.Status][row.Username] = struct{}{}
	}
	result.FailedUsersDaily = int64(len(failedUsers))
	result.SuccessUsersDaily = int64(len(successUsers))
	for loginType, statuses := range typeUsers {
		for status, users := range statuses {
			if status == "success" || status == "failed" {
				result.LoginTypeCounts = append(result.LoginTypeCounts, MetricsLoginTypeCount{LoginType: loginType, Status: status, Count: int64(len(users))})
			}
		}
	}
	sort.Slice(result.LoginTypeCounts, func(i, j int) bool {
		if result.LoginTypeCounts[i].LoginType == result.LoginTypeCounts[j].LoginType {
			return result.LoginTypeCounts[i].Status < result.LoginTypeCounts[j].Status
		}
		return result.LoginTypeCounts[i].LoginType < result.LoginTypeCounts[j].LoginType
	})

	operationRows, err := b.repo.OperationLogsSince(ctx, b.db, start)
	if err != nil {
		return nil, err
	}
	for _, row := range operationRows {
		result.OperationStatusDaily[row.Status]++
	}

	hostRows, err := b.repo.Hosts(ctx, b.db)
	if err != nil {
		return nil, err
	}
	sessions, err := b.repo.TerminalSessions(ctx, b.db)
	if err != nil {
		return nil, err
	}
	userIDs := make([]int64, 0)
	latestSession := make(map[int64]struct {
		UserID                        int64
		Username, SessionType, Status string
		At                            time.Time
		Has                           bool
	})
	for _, session := range sessions {
		if _, exists := latestSession[session.HostID]; exists {
			continue
		}
		latestSession[session.HostID] = struct {
			UserID                        int64
			Username, SessionType, Status string
			At                            time.Time
			Has                           bool
		}{UserID: session.UserID, SessionType: session.SessionType, At: session.StartedAt, Status: func() string {
			if session.EndedAt.IsZero() {
				return "active"
			}
			return "ended"
		}(), Has: true}
		if session.UserID > 0 {
			userIDs = append(userIDs, session.UserID)
		}
	}
	users, err := b.repo.Users(ctx, b.db, userIDs)
	if err != nil {
		return nil, err
	}
	usernames := make(map[int64]string)
	for _, user := range users {
		usernames[user.ID] = user.Username
	}
	for hostID, session := range latestSession {
		if session.Has {
			item := latestSession[hostID]
			item.Username = usernames[session.UserID]
			latestSession[hostID] = item
		}
	}
	for _, host := range hostRows {
		item := MetricsHost{ID: host.ID, Name: host.Name, IP: host.IP, OS: host.Os, Hostname: host.Hostname, Status: host.Status, CPUUsage: host.CPUUsage, MemoryUsage: host.MemoryUsage, DiskUsage: host.DiskUsage}
		if session, ok := latestSession[host.ID]; ok {
			item.RecentUser, item.RecentSessionType, item.RecentStatus, item.RecentAt, item.HasRecent = session.Username, session.SessionType, session.Status, session.At, true
		}
		result.Hosts = append(result.Hosts, item)
	}

	clusters, err := b.repo.KubernetesClusters(ctx, b.db)
	if err != nil {
		return nil, err
	}
	for _, cluster := range clusters {
		item := MetricsKubernetesCluster{ID: cluster.ID, Name: cluster.Name, Alias: cluster.Alias_, Status: cluster.Status, NodeCount: cluster.NodeCount, PodCount: cluster.PodCount}
		result.KubernetesClusters = append(result.KubernetesClusters, item)
		if cluster.Status == 1 {
			result.KubernetesHealthy++
			result.KubernetesNodesTotal += cluster.NodeCount
		}
	}

	result.MonitorDomainsTotal, err = b.repo.CountMonitorDomains(ctx, b.db)
	if err != nil {
		return nil, err
	}
	result.MonitorCertificatesValid, err = b.repo.CountValidMonitorCertificates(ctx, b.db)
	if err != nil {
		return nil, err
	}

	for _, days := range []int64{5, 15, 30, 60} {
		deadline := now.Add(time.Duration(days) * 24 * time.Hour)
		sslCount, err := b.repo.CountExpiringSSLCertificates(ctx, b.db, now, deadline)
		if err != nil {
			return nil, err
		}
		monitorCount, err := b.repo.CountExpiringMonitorCertificates(ctx, b.db, now, deadline)
		if err != nil {
			return nil, err
		}
		if result.CertificateExpiring[fmtInt(days)] == nil {
			result.CertificateExpiring[fmtInt(days)] = make(map[string]int64)
		}
		result.CertificateExpiring[fmtInt(days)]["ssl_certificate"] = sslCount
		result.CertificateExpiring[fmtInt(days)]["monitor_domain"] = monitorCount
	}
	return result, nil
}

func fmtInt(value int64) string { return fmt.Sprintf("%d", value) }
