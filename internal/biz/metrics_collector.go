package biz

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const metricsNamespace = "gops"

type MetricsCollector struct {
	metricsBiz *MetricsBiz

	recentLoginInfo           *prometheus.Desc
	recentLoginTimestamp      *prometheus.Desc
	failedUsersDaily          *prometheus.Desc
	successUsersDaily         *prometheus.Desc
	usersDailyByType          *prometheus.Desc
	operationStatusDaily      *prometheus.Desc
	operationStatusClassDaily *prometheus.Desc
	hostInfo                  *prometheus.Desc
	hostStatus                *prometheus.Desc
	hostCPUUsage              *prometheus.Desc
	hostMemoryUsage           *prometheus.Desc
	hostDiskUsage             *prometheus.Desc
	hostRecentLoginInfo       *prometheus.Desc
	hostRecentLoginTime       *prometheus.Desc
	kubernetesClusters        *prometheus.Desc
	kubernetesHealthyClusters *prometheus.Desc
	kubernetesNodesTotal      *prometheus.Desc
	kubernetesClusterNodes    *prometheus.Desc
	kubernetesClusterPods     *prometheus.Desc
	monitorDomainsTotal       *prometheus.Desc
	monitorCertificatesValid  *prometheus.Desc
	certificatesExpiring      *prometheus.Desc
	snapshotScrapeError       *prometheus.Desc
	snapshotScrapeTimestamp   *prometheus.Desc
}

func NewMetricsCollector(metricsBiz *MetricsBiz) *MetricsCollector {
	return &MetricsCollector{
		metricsBiz:                metricsBiz,
		recentLoginInfo:           prometheus.NewDesc(metricsNamespace+"_login_recent_info", "最近系统登录记录数量", []string{"login_type", "status"}, nil),
		recentLoginTimestamp:      prometheus.NewDesc(metricsNamespace+"_login_recent_timestamp_seconds", "最近系统登录时间戳", []string{"login_type", "status"}, nil),
		failedUsersDaily:          prometheus.NewDesc(metricsNamespace+"_login_failed_users_daily", "每日登录失败的去重用户数", []string{"date"}, nil),
		successUsersDaily:         prometheus.NewDesc(metricsNamespace+"_login_success_users_daily", "每日登录成功的去重用户数", []string{"date"}, nil),
		usersDailyByType:          prometheus.NewDesc(metricsNamespace+"_login_users_daily_by_type", "每日按登录类型统计的去重用户数", []string{"date", "login_type", "status"}, nil),
		operationStatusDaily:      prometheus.NewDesc(metricsNamespace+"_operation_requests_status_daily", "每日接口精确状态码请求数", []string{"date", "status_code"}, nil),
		operationStatusClassDaily: prometheus.NewDesc(metricsNamespace+"_operation_requests_status_class_daily", "每日接口状态码类别请求数", []string{"date", "status_class"}, nil),
		hostInfo:                  prometheus.NewDesc(metricsNamespace+"_host_info", "资产主机信息", []string{"host_id", "host_name", "host_ip", "os", "hostname"}, nil),
		hostStatus:                prometheus.NewDesc(metricsNamespace+"_host_status", "资产主机状态，1表示在线", []string{"host_id", "host_name", "host_ip"}, nil),
		hostCPUUsage:              prometheus.NewDesc(metricsNamespace+"_host_cpu_usage_percent", "资产主机CPU使用率", []string{"host_id", "host_name", "host_ip"}, nil),
		hostMemoryUsage:           prometheus.NewDesc(metricsNamespace+"_host_memory_usage_percent", "资产主机内存使用率", []string{"host_id", "host_name", "host_ip"}, nil),
		hostDiskUsage:             prometheus.NewDesc(metricsNamespace+"_host_disk_usage_percent", "资产主机磁盘使用率", []string{"host_id", "host_name", "host_ip"}, nil),
		hostRecentLoginInfo:       prometheus.NewDesc(metricsNamespace+"_host_recent_login_info", "资产主机最近终端登录信息", []string{"host_id", "host_name", "host_ip", "username", "session_type", "status"}, nil),
		hostRecentLoginTime:       prometheus.NewDesc(metricsNamespace+"_host_recent_login_timestamp_seconds", "资产主机最近终端登录时间戳", []string{"host_id", "host_name", "host_ip", "username", "session_type", "status"}, nil),
		kubernetesClusters:        prometheus.NewDesc(metricsNamespace+"_kubernetes_clusters_total", "容器管理集群总数", nil, nil),
		kubernetesHealthyClusters: prometheus.NewDesc(metricsNamespace+"_kubernetes_healthy_clusters", "运行正常的容器集群数", nil, nil),
		kubernetesNodesTotal:      prometheus.NewDesc(metricsNamespace+"_kubernetes_nodes_total", "容器管理集群节点总数", nil, nil),
		kubernetesClusterNodes:    prometheus.NewDesc(metricsNamespace+"_kubernetes_cluster_nodes", "每个容器集群的节点数", []string{"cluster_id", "cluster_name", "cluster_alias"}, nil),
		kubernetesClusterPods:     prometheus.NewDesc(metricsNamespace+"_kubernetes_cluster_pods", "每个容器集群的 Pod 数", []string{"cluster_id", "cluster_name", "cluster_alias"}, nil),
		monitorDomainsTotal:       prometheus.NewDesc(metricsNamespace+"_monitor_domains_total", "监控管理域名监控总个数", nil, nil),
		monitorCertificatesValid:  prometheus.NewDesc(metricsNamespace+"_monitor_certificates_valid", "监控管理证书状态正常的域名个数", nil, nil),
		certificatesExpiring:      prometheus.NewDesc(metricsNamespace+"_certificates_expiring", "未来指定天数内到期且未过期的证书数量", []string{"within_days", "source"}, nil),
		snapshotScrapeError:       prometheus.NewDesc(metricsNamespace+"_metrics_snapshot_scrape_error", "指标快照采集是否失败，1表示失败", nil, nil),
		snapshotScrapeTimestamp:   prometheus.NewDesc(metricsNamespace+"_metrics_snapshot_scrape_timestamp_seconds", "指标快照最近一次采集时间戳", nil, nil),
	}
}

func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.recentLoginInfo
	ch <- c.recentLoginTimestamp
	ch <- c.failedUsersDaily
	ch <- c.successUsersDaily
	ch <- c.usersDailyByType
	ch <- c.operationStatusDaily
	ch <- c.operationStatusClassDaily
	ch <- c.hostInfo
	ch <- c.hostStatus
	ch <- c.hostCPUUsage
	ch <- c.hostMemoryUsage
	ch <- c.hostDiskUsage
	ch <- c.hostRecentLoginInfo
	ch <- c.hostRecentLoginTime
	ch <- c.kubernetesClusters
	ch <- c.kubernetesHealthyClusters
	ch <- c.kubernetesNodesTotal
	ch <- c.kubernetesClusterNodes
	ch <- c.kubernetesClusterPods
	ch <- c.monitorDomainsTotal
	ch <- c.monitorCertificatesValid
	ch <- c.certificatesExpiring
	ch <- c.snapshotScrapeError
	ch <- c.snapshotScrapeTimestamp
}

func (c *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	now := time.Now()
	snapshot, err := c.metricsBiz.Snapshot(context.Background())
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.snapshotScrapeError, prometheus.GaugeValue, 1)
		ch <- prometheus.MustNewConstMetric(c.snapshotScrapeTimestamp, prometheus.GaugeValue, float64(now.Unix()))
		return
	}
	ch <- prometheus.MustNewConstMetric(c.snapshotScrapeError, prometheus.GaugeValue, 0)
	ch <- prometheus.MustNewConstMetric(c.snapshotScrapeTimestamp, prometheus.GaugeValue, float64(now.Unix()))

	date := now.Format("2006-01-02")
	recentLoginCounts := make(map[string]map[string]int64)
	recentLoginTimes := make(map[string]map[string]time.Time)
	for _, login := range snapshot.RecentLogins {
		loginType := login.LoginType
		if loginType == "" {
			loginType = "unknown"
		}
		if recentLoginCounts[loginType] == nil {
			recentLoginCounts[loginType] = make(map[string]int64)
			recentLoginTimes[loginType] = make(map[string]time.Time)
		}
		recentLoginCounts[loginType][login.Status]++
		if latest, exists := recentLoginTimes[loginType][login.Status]; !exists || login.CreatedAt.After(latest) {
			recentLoginTimes[loginType][login.Status] = login.CreatedAt
		}
	}
	for loginType, statuses := range recentLoginCounts {
		for status, count := range statuses {
			labels := []string{loginType, status}
			ch <- prometheus.MustNewConstMetric(c.recentLoginInfo, prometheus.GaugeValue, float64(count), labels...)
			ch <- prometheus.MustNewConstMetric(c.recentLoginTimestamp, prometheus.GaugeValue, float64(recentLoginTimes[loginType][status].Unix()), labels...)
		}
	}
	ch <- prometheus.MustNewConstMetric(c.failedUsersDaily, prometheus.GaugeValue, float64(snapshot.FailedUsersDaily), date)
	ch <- prometheus.MustNewConstMetric(c.successUsersDaily, prometheus.GaugeValue, float64(snapshot.SuccessUsersDaily), date)
	for _, item := range snapshot.LoginTypeCounts {
		ch <- prometheus.MustNewConstMetric(c.usersDailyByType, prometheus.GaugeValue, float64(item.Count), date, item.LoginType, item.Status)
	}

	classCounts := map[string]int64{}
	for status, count := range snapshot.OperationStatusDaily {
		ch <- prometheus.MustNewConstMetric(c.operationStatusDaily, prometheus.GaugeValue, float64(count), date, fmt.Sprintf("%d", status))
		if status >= 100 && status < 600 {
			classCounts[fmt.Sprintf("%dxx", status/100)] += count
		}
	}
	for _, className := range []string{"1xx", "2xx", "3xx", "4xx", "5xx"} {
		ch <- prometheus.MustNewConstMetric(c.operationStatusClassDaily, prometheus.GaugeValue, float64(classCounts[className]), date, className)
	}

	for _, host := range snapshot.Hosts {
		hostID := fmt.Sprintf("%d", host.ID)
		labels := []string{hostID, host.Name, host.IP}
		status := 0.0
		if host.Status == 1 {
			status = 1
		}
		ch <- prometheus.MustNewConstMetric(c.hostInfo, prometheus.GaugeValue, 1, hostID, host.Name, host.IP, host.OS, host.Hostname)
		ch <- prometheus.MustNewConstMetric(c.hostStatus, prometheus.GaugeValue, status, labels...)
		ch <- prometheus.MustNewConstMetric(c.hostCPUUsage, prometheus.GaugeValue, host.CPUUsage, labels...)
		ch <- prometheus.MustNewConstMetric(c.hostMemoryUsage, prometheus.GaugeValue, host.MemoryUsage, labels...)
		ch <- prometheus.MustNewConstMetric(c.hostDiskUsage, prometheus.GaugeValue, host.DiskUsage, labels...)
		if host.HasRecent {
			recentLabels := []string{hostID, host.Name, host.IP, host.RecentUser, host.RecentSessionType, host.RecentStatus}
			ch <- prometheus.MustNewConstMetric(c.hostRecentLoginInfo, prometheus.GaugeValue, 1, recentLabels...)
			ch <- prometheus.MustNewConstMetric(c.hostRecentLoginTime, prometheus.GaugeValue, float64(host.RecentAt.Unix()), recentLabels...)
		}
	}

	for _, cluster := range snapshot.KubernetesClusters {
		labels := []string{fmt.Sprintf("%d", cluster.ID), cluster.Name, cluster.Alias}
		if cluster.Status == 1 {
			ch <- prometheus.MustNewConstMetric(c.kubernetesClusterNodes, prometheus.GaugeValue, float64(cluster.NodeCount), labels...)
			ch <- prometheus.MustNewConstMetric(c.kubernetesClusterPods, prometheus.GaugeValue, float64(cluster.PodCount), labels...)
		} else {
			ch <- prometheus.MustNewConstMetric(c.kubernetesClusterNodes, prometheus.GaugeValue, 0, labels...)
			ch <- prometheus.MustNewConstMetric(c.kubernetesClusterPods, prometheus.GaugeValue, 0, labels...)
		}
	}
	ch <- prometheus.MustNewConstMetric(c.kubernetesClusters, prometheus.GaugeValue, float64(len(snapshot.KubernetesClusters)))
	ch <- prometheus.MustNewConstMetric(c.kubernetesHealthyClusters, prometheus.GaugeValue, float64(snapshot.KubernetesHealthy))
	ch <- prometheus.MustNewConstMetric(c.kubernetesNodesTotal, prometheus.GaugeValue, float64(snapshot.KubernetesNodesTotal))

	ch <- prometheus.MustNewConstMetric(c.monitorDomainsTotal, prometheus.GaugeValue, float64(snapshot.MonitorDomainsTotal))
	ch <- prometheus.MustNewConstMetric(c.monitorCertificatesValid, prometheus.GaugeValue, float64(snapshot.MonitorCertificatesValid))
	for days, sources := range snapshot.CertificateExpiring {
		for source, count := range sources {
			ch <- prometheus.MustNewConstMetric(c.certificatesExpiring, prometheus.GaugeValue, float64(count), days, source)
		}
	}
}
