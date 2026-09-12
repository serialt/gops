package biz

import (
	"errors"
	"strconv"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

type UserBiz struct {
	db     *gorm.DB
	casbin *casbinx.Service
}

func NewUserBiz(db *gorm.DB, casbinService *casbinx.Service) *UserBiz {
	return &UserBiz{db: db, casbin: casbinService}
}

func (b *UserBiz) List(req *api.UserListRequest) ([]*api.UserResponse, error) {
	var users []model.User
	query := b.db.Preload("Roles").Preload("Organizations").Order("id DESC")
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.OrganizationID > 0 {
		query = query.Joins("JOIN organization_user ON organization_user.user_id = \"user\".id").Where("organization_user.organization_id = ?", req.OrganizationID).Group("\"user\".id")
	}
	if err := query.Find(&users).Error; err != nil {
		return nil, err
	}
	result := make([]*api.UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, toUserResponse(user))
	}
	return result, nil
}

func (b *UserBiz) Create(req *api.UserRequest) (*api.UserResponse, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := model.User{Username: req.Username, PasswordHash: hash, Nickname: req.Nickname, Status: *req.Status}
	if err := b.saveUser(&user, req.RoleIDs, req.OrganizationIDs, true); err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (b *UserBiz) Update(req *api.UserRequest) (*api.UserResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var user model.User
	if err := b.db.Preload("Roles").Preload("Organizations").First(&user, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	user.Username = req.Username
	user.Nickname = req.Nickname
	user.Status = *req.Status
	if req.Password != "" {
		hash, err := auth.HashPassword(req.Password)
		if err != nil {
			return nil, err
		}
		user.PasswordHash = hash
	}
	if err := b.saveUser(&user, req.RoleIDs, req.OrganizationIDs, false); err != nil {
		return nil, err
	}
	return toUserResponse(user), nil
}

func (b *UserBiz) Delete(req *api.UserIDRequest, currentID uint) error {
	if int64(currentID) == req.ID {
		return errors.New("cannot delete current user")
	}
	if err := b.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", req.ID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		result := tx.Delete(&model.User{}, req.ID)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}
		return nil
	}); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if _, err := b.casbin.Enforcer.DeleteUser(strconv.FormatInt(req.ID, 10)); err != nil {
		return err
	}
	return nil
}

func (b *UserBiz) saveUser(user *model.User, roleIDs []int64, organizationIDs []int64, create bool) error {
	var roles []*model.Role
	if len(roleIDs) > 0 {
		if err := b.db.Where("id IN ?", roleIDs).Find(&roles).Error; err != nil {
			return err
		}
		if len(roles) != len(roleIDs) {
			return errors.New("one or more roles do not exist")
		}
	}
	var organizations []*model.Organization
	if len(organizationIDs) > 0 {
		if err := b.db.Where("id IN ?", organizationIDs).Find(&organizations).Error; err != nil {
			return err
		}
		if len(organizations) != len(organizationIDs) {
			return errors.New("one or more organizations do not exist")
		}
	}
	var root model.Organization
	if err := b.db.Where("code = ?", "root").First(&root).Error; err != nil {
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

	if err := b.db.Transaction(func(tx *gorm.DB) error {
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
	if err := b.casbin.SetUserRoles(uint(user.ID), codes); err != nil {
		return err
	}
	user.Roles = roles
	user.Organizations = organizations
	return nil
}

func toUserResponse(user model.User) *api.UserResponse {
	roles := make([]*api.RoleResponse, 0, len(user.Roles))
	for _, role := range user.Roles {
		if role != nil {
			roles = append(roles, toRoleResponse(*role))
		}
	}
	organizations := make([]*api.OrganizationResponse, 0, len(user.Organizations))
	for _, organization := range user.Organizations {
		if organization != nil {
			organizations = append(organizations, toOrganizationResponse(*organization))
		}
	}
	return &api.UserResponse{ID: user.ID, Username: user.Username, Nickname: user.Nickname, Status: user.Status, Roles: roles, Organizations: organizations, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}
