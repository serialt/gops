package api

import "time"

type AssetGroupListRequest struct {
	Keyword string `json:"keyword"`
}
type AssetGroupRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Code        string `json:"code" binding:"required"`
	ParentID    int64  `json:"parent_id"`
	Description string `json:"description"`
	SortOrder   int64  `json:"sort_order"`
	Status      *bool  `json:"status" binding:"required"`
}
type AssetGroupIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type AssetGroupResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	ParentID    int64     `json:"parent_id"`
	Description string    `json:"description"`
	SortOrder   int64     `json:"sort_order"`
	Status      bool      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CredentialListRequest struct {
	Keyword string `json:"keyword"`
}
type CredentialRequest struct {
	ID             int64  `json:"id"`
	Name           string `json:"name" binding:"required"`
	CredentialType string `json:"credential_type" binding:"required"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	PrivateKey     string `json:"private_key"`
	Passphrase     string `json:"passphrase"`
	Description    string `json:"description"`
}
type CredentialIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type CredentialResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	CredentialType string    `json:"credential_type"`
	Username       string    `json:"username"`
	Description    string    `json:"description"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type CloudAccountListRequest struct {
	Keyword string `json:"keyword"`
}
type CloudProviderOptionResponse struct {
	Value   string                       `json:"value"`
	Label   string                       `json:"label"`
	Regions []*CloudRegionOptionResponse `json:"regions"`
}
type CloudRegionOptionResponse struct {
	Value string `json:"value"`
	Label string `json:"label"`
}
type CloudAccountRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Provider    string `json:"provider" binding:"required"`
	AccessKey   string `json:"access_key"`
	SecretKey   string `json:"secret_key"`
	Region      string `json:"region"`
	Description string `json:"description"`
	Status      *bool  `json:"status" binding:"required"`
}
type CloudAccountIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type CloudAccountResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Provider    string    `json:"provider"`
	Region      string    `json:"region"`
	Description string    `json:"description"`
	Status      bool      `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// DatabaseTypeOptionResponse represents one database type and its default port.
type DatabaseTypeOptionResponse struct {
	Value       string `json:"value"`
	Label       string `json:"label"`
	DefaultPort int64  `json:"default_port"`
}

type HostListRequest struct {
	Keyword string `json:"keyword"`
	GroupID int64  `json:"group_id"`
}
type HostRequest struct {
	ID             int64  `json:"id"`
	Name           string `json:"name" binding:"required"`
	GroupID        int64  `json:"group_id"`
	HostType       string `json:"host_type"`
	IP             string `json:"ip" binding:"required"`
	Port           int64  `json:"port" binding:"min=1,max=65535"`
	SSHUser        string `json:"ssh_user"`
	CredentialID   int64  `json:"credential_id"`
	CloudAccountID int64  `json:"cloud_account_id"`
	Tags           string `json:"tags"`
	Description    string `json:"description"`
	Status         int64  `json:"status"`
}
type HostIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type HostResponse struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`
	GroupID        int64     `json:"group_id"`
	HostType       string    `json:"host_type"`
	IP             string    `json:"ip"`
	Port           int64     `json:"port"`
	SSHUser        string    `json:"ssh_user"`
	CredentialID   int64     `json:"credential_id"`
	CloudAccountID int64     `json:"cloud_account_id"`
	Tags           string    `json:"tags"`
	Description    string    `json:"description"`
	Status         int64     `json:"status"`
	LastSeenAt     time.Time `json:"last_seen_at"`
	OS             string    `json:"os"`
	Kernel         string    `json:"kernel"`
	Arch           string    `json:"arch"`
	CPUInfo        string    `json:"cpu_info"`
	CPUCores       int64     `json:"cpu_cores"`
	CPUUsage       float64   `json:"cpu_usage"`
	MemoryTotal    int64     `json:"memory_total"`
	MemoryUsed     int64     `json:"memory_used"`
	MemoryUsage    float64   `json:"memory_usage"`
	DiskTotal      int64     `json:"disk_total"`
	DiskUsed       int64     `json:"disk_used"`
	DiskUsage      float64   `json:"disk_usage"`
	Uptime         string    `json:"uptime"`
	Hostname       string    `json:"hostname"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DbInstanceListRequest struct {
	Keyword string `json:"keyword"`
	GroupID int64  `json:"group_id"`
}
type DbInstanceRequest struct {
	ID           int64  `json:"id"`
	GroupID      int64  `json:"group_id"`
	Name         string `json:"name" binding:"required"`
	DbType       string `json:"db_type" binding:"required"`
	Host         string `json:"host" binding:"required"`
	Port         int64  `json:"port" binding:"min=1,max=65535"`
	Version      string `json:"version"`
	CredentialID int64  `json:"credential_id"`
	Tags         string `json:"tags"`
	Description  string `json:"description"`
	Status       *bool  `json:"status" binding:"required"`
}
type DbInstanceIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type DbInstanceTestRequest struct {
	DbType       string `json:"db_type" binding:"required"`
	Host         string `json:"host" binding:"required"`
	Port         int64  `json:"port" binding:"min=1,max=65535"`
	CredentialID int64  `json:"credential_id" binding:"required"`
}
type DbInstanceTestResponse struct {
	Version string `json:"version"`
}
type DbInstanceResponse struct {
	ID           int64     `json:"id"`
	GroupID      int64     `json:"group_id"`
	Name         string    `json:"name"`
	DbType       string    `json:"db_type"`
	Host         string    `json:"host"`
	Port         int64     `json:"port"`
	Version      string    `json:"version"`
	CredentialID int64     `json:"credential_id"`
	Tags         string    `json:"tags"`
	Description  string    `json:"description"`
	Status       bool      `json:"status"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type TerminalSessionListRequest struct {
	Keyword     string `json:"keyword"`
	SessionType string `json:"session_type"`
}
type TerminalSessionIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type TerminalSessionResponse struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	HostID          int64     `json:"host_id"`
	DbInstanceID    int64     `json:"db_instance_id"`
	SessionType     string    `json:"session_type"`
	StartedAt       time.Time `json:"started_at"`
	EndedAt         time.Time `json:"ended_at"`
	DurationSeconds int64     `json:"duration_seconds"`
	CommandCount    int64     `json:"command_count"`
}
type TerminalSessionRecordingResponse struct {
	ID        int64  `json:"id"`
	Recording string `json:"recording"`
}
