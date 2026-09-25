package api

type DashboardStatsResponse struct {
	HostsTotal      int64 `json:"hosts_total"`
	HostsOnline     int64 `json:"hosts_online"`
	HostsOffline    int64 `json:"hosts_offline"`
	ClustersTotal   int64 `json:"clusters_total"`
	NodesTotal      int64 `json:"nodes_total"`
	PodsTotal       int64 `json:"pods_total"`
	TodayOperations int64 `json:"today_operations"`
	ActiveAlerts    int64 `json:"active_alerts"`
}
