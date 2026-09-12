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
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type DbInstanceListRequest struct {
	Keyword string `json:"keyword"`
}
type DbInstanceRequest struct {
	ID           int64  `json:"id"`
	Name         string `json:"name" binding:"required"`
	DbType       string `json:"db_type" binding:"required"`
	Host         string `json:"host" binding:"required"`
	Port         int64  `json:"port" binding:"min=1,max=65535"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	Version      string `json:"version"`
	CredentialID int64  `json:"credential_id"`
	Tags         string `json:"tags"`
	Description  string `json:"description"`
	Status       *bool  `json:"status" binding:"required"`
}
type DbInstanceIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}
type DbInstanceResponse struct {
	ID           int64     `json:"id"`
	Name         string    `json:"name"`
	DbType       string    `json:"db_type"`
	Host         string    `json:"host"`
	Port         int64     `json:"port"`
	Username     string    `json:"username"`
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
