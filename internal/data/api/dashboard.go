package api

import "time"

type DashboardDomainCertificateResponse struct {
	Domain    string    `json:"domain"`
	SSLExpiry time.Time `json:"ssl_expiry"`
}

type DashboardStatsResponse struct {
	HostsTotal         int64                                 `json:"hosts_total"`
	HostsOnline        int64                                 `json:"hosts_online"`
	HostsOffline       int64                                 `json:"hosts_offline"`
	ClustersTotal      int64                                 `json:"clusters_total"`
	ClustersOffline    int64                                 `json:"clusters_offline"`
	NodesTotal         int64                                 `json:"nodes_total"`
	NodesOffline       int64                                 `json:"nodes_offline"`
	PodsTotal          int64                                 `json:"pods_total"`
	PodsAbnormal       int64                                 `json:"pods_abnormal"`
	DomainsTotal       int64                                 `json:"domains_total"`
	DomainsNormal      int64                                 `json:"domains_normal"`
	DomainsAbnormal    int64                                 `json:"domains_abnormal"`
	DomainsPaused      int64                                 `json:"domains_paused"`
	DomainCertificates []*DashboardDomainCertificateResponse `json:"domain_certificates"`
	TodayOperations    int64                                 `json:"today_operations"`
	ActiveAlerts       int64                                 `json:"active_alerts"`
}
