package api

import "time"

type RoleRequest struct {
	ID          uint   `json:"id"`
	Code        string `json:"code" binding:"required"`
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
}

type RoleIDRequest struct {
	ID uint `json:"id" binding:"required"`
}

type PermissionRequest struct {
	ID          uint         `json:"id" binding:"required"`
	Permissions []Permission `json:"permissions"`
}

type Permission struct {
	Group  string `json:"group"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Method string `json:"method"`
	Key    string `json:"key"`
}

type RoleResponse struct {
	ID          int64     `json:"id"`
	Code        string    `json:"code"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
