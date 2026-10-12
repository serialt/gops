package define

import "time"

type PageCond struct {
	Page  int
	Limit int
}

type AssetGroupListCond struct {
	Page     int
	Limit    int
	Keyword  string
	ParentID int64
}

type CredentialListCond struct {
	PageCond
	Keyword string
}

type CloudAccountListCond struct {
	PageCond
	Keyword string
}

type HostListCond struct {
	PageCond
	Keyword string
	GroupID int64
	Status  *int64
	IDs     []int64
}

type HostCollectedInfo struct {
	LastSeenAt  time.Time
	OS          string
	Kernel      string
	Arch        string
	CPUInfo     string
	CPUCores    int64
	CPUUsage    float64
	MemoryTotal int64
	MemoryUsed  int64
	MemoryUsage float64
	DiskTotal   int64
	DiskUsed    int64
	DiskUsage   float64
	Uptime      string
	Hostname    string
}

type DBInstanceListCond struct {
	PageCond
	Keyword string
	GroupID int64
	DbType  string
	IDs     []int64
}
