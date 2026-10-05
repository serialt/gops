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
type IOTACLRuleRequest struct {
	ID          int64  `json:"id"`
	RoleID      int64  `json:"role_id" binding:"required,min=1"`
	Operation   string `json:"operation" binding:"required,oneof=publish subscribe"`
	Effect      string `json:"effect" binding:"required,oneof=allow deny"`
	TopicFilter string `json:"topic_filter" binding:"required"`
}
type IOTPublishRequest struct {
	Broker  string `json:"broker"`
	Topic   string `json:"topic" binding:"required"`
	Payload string `json:"payload"`
	QoS     int64  `json:"qos" binding:"min=0,max=2"`
	Retain  bool   `json:"retain"`
}

type IOTHTTPReceiveRequest struct {
	Source   string `json:"source" binding:"required"`
	ClientID string `json:"client_id"`
	Topic    string `json:"topic" binding:"required"`
	Payload  string `json:"payload"`
	QoS      int64  `json:"qos" binding:"min=0,max=2"`
	Retain   bool   `json:"retain"`
}

type IOTExternalBrokerListRequest struct {
	Page    int64  `json:"page" binding:"required,min=1"`
	Limit   int64  `json:"limit" binding:"required,min=1,max=1000"`
	Keyword string `json:"keyword"`
	Enabled *bool  `json:"enabled"`
}

type IOTExternalBrokerRequest struct {
	ID                 int64  `json:"id"`
	Name               string `json:"name" binding:"required"`
	URL                string `json:"url" binding:"required"`
	ClientID           string `json:"client_id" binding:"required"`
	Username           string `json:"username"`
	Password           string `json:"password"`
	TopicFilter        string `json:"topic_filter" binding:"required"`
	Enabled            bool   `json:"enabled"`
	InsecureSkipVerify bool   `json:"insecure_skip_verify"`
}

type IOTExternalBrokerResponse struct {
	ID                 int64     `json:"id"`
	Name               string    `json:"name"`
	URL                string    `json:"url"`
	ClientID           string    `json:"client_id"`
	Username           string    `json:"username"`
	TopicFilter        string    `json:"topic_filter"`
	Enabled            bool      `json:"enabled"`
	Connected          bool      `json:"connected"`
	HasPassword        bool      `json:"has_password"`
	InsecureSkipVerify bool      `json:"insecure_skip_verify"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

type IOTBrokerOptionResponse struct {
	Key       string `json:"key"`
	Name      string `json:"name"`
	URL       string `json:"url"`
	Type      string `json:"type"`
	Connected bool   `json:"connected"`
	Internal  bool   `json:"internal"`
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
type IOTDashboardResponse struct {
	TopicTotal     int64                 `json:"topic_total"`
	MessageTotal   int64                 `json:"message_total"`
	ClientTotal    int64                 `json:"client_total"`
	OnlineTotal    int64                 `json:"online_total"`
	RecentMessages []*IOTMessageResponse `json:"recent_messages"`
	RecentClients  []*IOTClientResponse  `json:"recent_clients"`
}
