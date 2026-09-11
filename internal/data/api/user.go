package api

import "time"

type UserListRequest struct {
	Keyword        string `json:"keyword"`
	OrganizationID int64  `json:"organization_id"`
}

type UserRequest struct {
	ID              int64   `json:"id"`
	Username        string  `json:"username" binding:"required"`
	Password        string  `json:"password"`
	Nickname        string  `json:"nickname" binding:"required"`
	Status          *bool   `json:"status" binding:"required"`
	RoleIDs         []int64 `json:"role_ids"`
	OrganizationIDs []int64 `json:"organization_ids"`
}

type UserIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type UserResponse struct {
	ID            int64                   `json:"id"`
	Username      string                  `json:"username"`
	Nickname      string                  `json:"nickname"`
	Status        bool                    `json:"status"`
	Roles         []*RoleResponse         `json:"roles"`
	Organizations []*OrganizationResponse `json:"organizations"`
	CreatedAt     time.Time               `json:"created_at"`
	UpdatedAt     time.Time               `json:"updated_at"`
}

type CurrentUserResponse struct {
	User        *UserResponse `json:"user"`
	Permissions []string      `json:"permissions"`
}
