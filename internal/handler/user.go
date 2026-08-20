package handler

import (
	"errors"
	"strconv"

	"gops/internal/data/model"
	"gops/internal/middleware"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"
	"gops/internal/pkg/response"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db     *gorm.DB
	casbin *casbinx.Service
}

func NewUserHandler(db *gorm.DB, casbinService *casbinx.Service) *UserHandler {
	return &UserHandler{db: db, casbin: casbinService}
}

type userRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password"`
	Nickname string `json:"nickname" binding:"required"`
	Status   *bool  `json:"status" binding:"required"`
	RoleIDs  []uint `json:"role_ids"`
}

func (h *UserHandler) List(c *gin.Context) {
	var users []model.User
	query := h.db.Preload("Roles").Order("id DESC")
	if keyword := c.Query("keyword"); keyword != "" {
		query = query.Where("username ILIKE ? OR nickname ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}
	if err := query.Find(&users).Error; err != nil {
		response.Error(c, 500, "query users failed")
		return
	}
	response.OK(c, users)
}

func (h *UserHandler) Create(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil || req.Password == "" {
		response.Error(c, 400, "invalid user data")
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		response.Error(c, 500, "hash password failed")
		return
	}
	user := model.User{Username: req.Username, PasswordHash: hash, Nickname: req.Nickname, Status: *req.Status}

	if err := h.saveUser(&user, req.RoleIDs, true); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "invalid user id")
		return
	}
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid user data")
		return
	}

	var user model.User
	if err := h.db.First(&user, uint(id)).Error; err != nil {
		response.Error(c, 404, "user not found")
		return
	}
	user.Username = req.Username
	user.Nickname = req.Nickname
	user.Status = *req.Status
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			response.Error(c, 500, "hash password failed")
			return
		}
		user.PasswordHash = hash
	}

	if err := h.saveUser(&user, req.RoleIDs, false); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		response.Error(c, 400, "invalid user id")
		return
	}
	currentID, _ := middleware.CurrentUserID(c)
	if currentID == uint(id) {
		response.Error(c, 400, "cannot delete current user")
		return
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", uint(id)).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.User{}, uint(id))
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			response.Error(c, 404, "user not found")
			return
		}
		response.Error(c, 500, "delete user failed")
		return
	}
	if _, err := h.casbin.Enforcer.DeleteUser(strconv.FormatUint(id, 10)); err != nil {
		response.Error(c, 500, "delete user policy failed")
		return
	}
	response.OK(c, nil)
}

func (h *UserHandler) saveUser(user *model.User, roleIDs []uint, create bool) error {
	var roles []*model.Role
	if len(roleIDs) > 0 {
		if err := h.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return errors.New("one or more roles do not exist")
		}
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if create {
			if err := tx.Create(user).Error; err != nil {
				return err
			}
		} else if err := tx.Save(user).Error; err != nil {
			return err
		}
		return tx.Model(user).Association("Roles").Replace(roles)
	}); err != nil {
		return err
	}

	codes := make([]string, 0, len(roles))
	for _, role := range roles {
		codes = append(codes, role.Code)
	}
	if err := h.casbin.SetUserRoles(uint(user.ID), codes); err != nil {
		return err
	}
	user.Roles = roles
	return nil
}
