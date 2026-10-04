package api

import "time"

type IOTListRequest struct {
	Page    int64  `json:"page" binding:"required,min=1"`
	Limit   int64  `json:"limit" binding:"required,min=1,max=1000"`
	Keyword string `json:"keyword"`
	Status  string `json:"status"`
	RoleID  int64  `json:"role_id"`
}

type IOTIDRequest struct {
	ID int64 `json:"id" binding:"required,min=1"`
}
type IOTUserRequest struct {
	ID       int64   `json:"id"`
	Username string  `json:"username" binding:"required"`
	Password string  `json:"password"`
	Enabled  bool    `json:"enabled"`
	RoleIDs  []int64 `json:"role_ids"`
}
type IOTRoleRequest struct {
	ID          int64  `json:"id"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}
type IOTACLRuleRequest struct {
	ID          int64  `json:"id"`
	RoleID      int64  `json:"role_id" binding:"required,min=1"`
	Operation   string `json:"operation" binding:"required,oneof=publish subscribe"`
	Effect      string `json:"effect" binding:"required,oneof=allow deny"`
	TopicFilter string `json:"topic_filter" binding:"required"`
}
type IOTPublishRequest struct {
	Topic   string `json:"topic" binding:"required"`
	Payload string `json:"payload"`
	QoS     int64  `json:"qos" binding:"min=0,max=2"`
	Retain  bool   `json:"retain"`
}

type IOTRoleResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
type IOTUserResponse struct {
	ID        int64              `json:"id"`
	Username  string             `json:"username"`
	Enabled   bool               `json:"enabled"`
	Roles     []*IOTRoleResponse `json:"roles"`
	CreatedAt time.Time          `json:"created_at"`
	UpdatedAt time.Time          `json:"updated_at"`
}
type IOTACLRuleResponse struct {
	ID          int64     `json:"id"`
	RoleID      int64     `json:"role_id"`
	RoleName    string    `json:"role_name"`
	Operation   string    `json:"operation"`
	Effect      string    `json:"effect"`
	TopicFilter string    `json:"topic_filter"`
	CreatedAt   time.Time `json:"created_at"`
}
type IOTTopicResponse struct {
	ID        int64     `json:"id"`
	Topic     string    `json:"topic"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type IOTMessageResponse struct {
	ID         int64     `json:"id"`
	ReceivedAt time.Time `json:"received_at"`
	ClientID   string    `json:"client_id"`
	Username   string    `json:"username"`
	Topic      string    `json:"topic"`
	Payload    string    `json:"payload"`
	QoS        int64     `json:"qos"`
	Retained   bool      `json:"retained"`
	Duplicate  bool      `json:"duplicate"`
}
type IOTClientResponse struct {
	ID               int64      `json:"id"`
	ClientID         string     `json:"client_id"`
	Status           string     `json:"status"`
	Username         string     `json:"username"`
	RemoteAddress    string     `json:"remote_address"`
	Listener         string     `json:"listener"`
	ProtocolVersion  int64      `json:"protocol_version"`
	CleanSession     bool       `json:"clean_session"`
	ConnectedAt      time.Time  `json:"connected_at"`
	DisconnectedAt   *time.Time `json:"disconnected_at"`
	DisconnectReason string     `json:"disconnect_reason"`
	SessionExpired   bool       `json:"session_expired"`
}
