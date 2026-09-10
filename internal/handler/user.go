package handler

import (
	"errors"

	"gops/internal/biz"
	"gops/internal/data/api"
	"gops/internal/middleware"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	biz *biz.UserBiz
}

func NewUserHandler(userBiz *biz.UserBiz) *UserHandler {
	return &UserHandler{biz: userBiz}
}

// List POST /api/v1/users/list 查询用户列表。
func (h *UserHandler) List(c *gin.Context) {
	req := new(api.UserListRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid user query")
		return
	}
	result, err := h.biz.List(req)
	if err != nil {
		response.Error(c, 500, "query users failed")
		return
	}
	response.OK(c, result)
}

// Create POST /api/v1/users/create 创建用户。
func (h *UserHandler) Create(c *gin.Context) {
	req := new(api.UserRequest)
	if err := c.ShouldBindJSON(req); err != nil || req.Password == "" {
		response.Error(c, 400, "invalid user data")
		return
	}
	result, err := h.biz.Create(req)
	if err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, result)
}

// Update POST /api/v1/users/update 更新用户。
func (h *UserHandler) Update(c *gin.Context) {
	req := new(api.UserRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid user data")
		return
	}
	result, err := h.biz.Update(req)
	if err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "user not found")
			return
		}
		if errors.Is(err, biz.ErrInvalid) {
			response.Error(c, 400, "invalid user id")
			return
		}
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, result)
}

// Delete POST /api/v1/users/delete 删除用户。
func (h *UserHandler) Delete(c *gin.Context) {
	req := new(api.UserIDRequest)
	if err := c.ShouldBindJSON(req); err != nil {
		response.Error(c, 400, "invalid user id")
		return
	}
	currentID, _ := middleware.CurrentUserID(c)
	if err := h.biz.Delete(req, currentID); err != nil {
		if errors.Is(err, biz.ErrNotFound) {
			response.Error(c, 404, "user not found")
			return
		}
		if err.Error() == "cannot delete current user" {
			response.Error(c, 400, err.Error())
			return
		}
		response.Error(c, 500, err.Error())
		return
	}
	response.OK(c, nil)
}
