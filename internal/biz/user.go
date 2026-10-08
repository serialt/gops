package biz

import (
	"context"
	"errors"
	"strconv"
	"strings"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/data/repo/rbac"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

type UserBiz struct {
	db     *gorm.DB
	casbin *casbinx.Service
	repo   *rbac.RbacRepo
}

func NewUserBiz(db *gorm.DB, casbinService *casbinx.Service, repo *rbac.RbacRepo) *UserBiz {
	return &UserBiz{db: db, casbin: casbinService, repo: repo}
}

func (b *UserBiz) ListForCLI(ctx context.Context) ([]*api.UserResponse, error) {
	users, err := b.repo.ListUsers(ctx, b.db)
	if err != nil {
		return nil, err
	}
	result := make([]*api.UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, toUserResponse(*user))
	}
	return result, nil
}

func (b *UserBiz) CreateForCLI(ctx context.Context, username, password, nickname string) (*api.UserResponse, error) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, errors.New("username and password are required")
	}
	if nickname = strings.TrimSpace(nickname); nickname == "" {
		nickname = username
	}
	if _, err := b.repo.GetUserByUsername(ctx, b.db, username); err == nil {
		return nil, errors.New("username already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}
	user := &model.User{Username: username, PasswordHash: hash, Nickname: nickname, Status: true}
	err = b.db.Transaction(func(tx *gorm.DB) error {
		root, err := b.repo.GetOrganizationByCode(ctx, tx, "root")
		if err != nil {
			return err
		}
		adminRole, err := b.repo.GetRoleByCode(ctx, tx, "admin")
		if err != nil {
			return err
		}
		if err := b.repo.CreateUser(ctx, tx, user); err != nil {
			return err
		}
		if err := b.repo.AppendUserOrganization(ctx, tx, user, root); err != nil {
			return err
		}
		return b.repo.AppendUserRole(ctx, tx, user, adminRole)
	})
	if err != nil {
		return nil, err
	}
	if err := b.casbin.SetUserRoles(uint(user.ID), []string{"admin"}); err != nil {
		return nil, err
	}
	return toUserResponse(*user), nil
}

func (b *UserBiz) ResetPasswordForCLI(ctx context.Context, username, password string) error {
	if strings.TrimSpace(username) == "" || password == "" {
		return errors.New("username and password are required")
	}
	user, err := b.repo.GetUserByUsername(ctx, b.db, strings.TrimSpace(username))
	if err != nil {
		return err
	}
	hash, err := auth.HashPassword(password)
	if err != nil {
		return err
	}
	return b.repo.UpdateUserPassword(ctx, b.db, user.ID, hash)
}

func (b *UserBiz) List(req *api.UserListRequest) (*api.PageResponse[*api.UserResponse], error) {
	query := b.db.Model(&model.User{}).Preload("Roles").Preload("Organizations").Order("id DESC")
	if req.Keyword != "" {
		query = query.Where("username LIKE ? OR nickname LIKE ? OR email LIKE ?", "%"+req.Keyword+"%", "%"+req.Keyword+"%", "%"+req.Keyword+"%")
	}
	if req.OrganizationID > 0 {
		query = query.Where("EXISTS (SELECT 1 FROM organization_user WHERE organization_user.user_id = \"user\".id AND organization_user.organization_id = ?)", req.OrganizationID)
	}
	users, total, err := queryPage[model.User](query, req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.UserResponse, 0, len(users))
	for _, user := range users {
		result = append(result, toUserResponse(user))
	}
	return &api.PageResponse[*api.UserResponse]{List: result, Total: total}, nil
}

func (b *UserBiz) Create(req *api.UserRequest) (*api.UserResponse, error) {
	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}
	user := model.User{Username: req.Username, Email: strings.ToLower(strings.TrimSpace(req.Email)), PasswordHash: hash, Nickname: req.Nickname, Status: *req.Status}
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
	user.Email = strings.ToLower(strings.TrimSpace(req.Email))
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
	return &api.UserResponse{ID: user.ID, Username: user.Username, Email: user.Email, Nickname: user.Nickname, Status: user.Status, Roles: roles, Organizations: organizations, CreatedAt: user.CreatedAt, UpdatedAt: user.UpdatedAt}
}
