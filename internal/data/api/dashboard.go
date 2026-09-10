package api

type DashboardStatsResponse struct {
	Users        int64 `json:"users"`
	EnabledUsers int64 `json:"enabled_users"`
	Roles        int64 `json:"roles"`
	Policies     int64 `json:"policies"`
}
