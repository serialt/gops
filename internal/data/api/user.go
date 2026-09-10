package api

import "time"

type UserListRequest struct {
	Keyword        string `json:"keyword"`
	OrganizationID uint   `json:"organization_id"`
}

type UserRequest struct {
	ID              uint   `json:"id"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password"`
	Nickname        string `json:"nickname" binding:"required"`
	Status          *bool  `json:"status" binding:"required"`
	RoleIDs         []uint `json:"role_ids"`
	OrganizationIDs []uint `json:"organization_ids"`
}

type UserIDRequest struct {
	ID uint `json:"id" binding:"required"`
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
