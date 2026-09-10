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
	ID              uint   `json:"id"`
	Username        string `json:"username" binding:"required"`
	Password        string `json:"password"`
	Nickname        string `json:"nickname" binding:"required"`
	Status          *bool  `json:"status" binding:"required"`
	RoleIDs         []uint `json:"role_ids"`
	OrganizationIDs []uint `json:"organization_ids"`
}
type userListRequest struct {
	Keyword        string `json:"keyword"`
	OrganizationID uint   `json:"organization_id"`
}
type idRequest struct {
	ID uint `json:"id" binding:"required"`
}

func (h *UserHandler) List(c *gin.Context) {
	var req userListRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid user query")
		return
	}
	var users []model.User
	query := h.db.Preload("Roles").Preload("Organizations").Order("id DESC")
	if req.Keyword != "" {
		query = query.Where("username ILIKE ? OR nickname ILIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.OrganizationID > 0 {
		query = query.Joins("JOIN organization_user ON organization_user.user_id = \"user\".id").Where("organization_user.organization_id = ?", req.OrganizationID).Group("\"user\".id")
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

	if err := h.saveUser(&user, req.RoleIDs, req.OrganizationIDs, true); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.Created(c, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	var req userRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid user data")
		return
	}

	var user model.User
	if req.ID == 0 {
		response.Error(c, 400, "invalid user id")
		return
	}
	if err := h.db.First(&user, req.ID).Error; err != nil {
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

	if err := h.saveUser(&user, req.RoleIDs, req.OrganizationIDs, false); err != nil {
		response.Error(c, 400, err.Error())
		return
	}
	response.OK(c, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	var req idRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "invalid user id")
		return
	}
	id := uint64(req.ID)
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

func (h *UserHandler) saveUser(user *model.User, roleIDs []uint, organizationIDs []uint, create bool) error {
	var roles []*model.Role
	if len(roleIDs) > 0 {
		if err := h.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return errors.New("one or more roles do not exist")
		}
	}
	var organizations []*model.Organization
	if len(organizationIDs) > 0 {
		if err := h.db.Where("id IN ?", organizationIDs).Find(&organizations).Error; err != nil {
			return err
		}
		if len(organizations) != len(organizationIDs) {
			return errors.New("one or more organizations do not exist")
		}
	}
	var root model.Organization
	if err := h.db.Where("code = ?", "root").First(&root).Error; err != nil {
		return errors.New("default root organization not found")
	}
	hasRoot := false
	for _, organization := range organizations {
		if organization.ID == root.ID {
			hasRoot = true
			break
		}
	}
	if !hasRoot {
		organizations = append(organizations, &root)
	}

	if err := h.db.Transaction(func(tx *gorm.DB) error {
		if create {
			if err := tx.Create(user).Error; err != nil {
				return err
			}
		} else if err := tx.Save(user).Error; err != nil {
			return err
		}
		if err := tx.Model(user).Association("Roles").Replace(roles); err != nil {
			return err
		}
		return tx.Model(user).Association("Organizations").Replace(organizations)
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
	user.Organizations = organizations
	return nil
}
