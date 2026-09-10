package api

import "time"

type OrganizationQueryRequest struct {
	Keyword string `json:"keyword"`
}

type OrganizationRequest struct {
	ID       int64  `json:"id"`
	ParentID *int64 `json:"parent_id"`
	Name     string `json:"name" binding:"required"`
	Code     string `json:"code" binding:"required"`
	Manager  string `json:"manager"`
	Status   *bool  `json:"status" binding:"required"`
}

type OrganizationMoveRequest struct {
	ID        int64  `json:"id" binding:"required"`
	ParentID  *int64 `json:"parent_id"`
	SortOrder int64  `json:"sort_order"`
}

type OrganizationIDRequest struct {
	ID int64 `json:"id" binding:"required"`
}

type OrganizationResponse struct {
	ID        int64     `json:"id"`
	ParentID  *int64    `json:"parent_id"`
	SortOrder int64     `json:"sort_order"`
	Name      string    `json:"name"`
	Code      string    `json:"code"`
	Manager   string    `json:"manager"`
	Status    bool      `json:"status"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
