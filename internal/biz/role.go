package biz

import (
	"errors"

	"gops/internal/data/api"
	"gops/internal/data/model"
	"gops/internal/pkg/casbinx"

	"gorm.io/gorm"
)

type RoleBiz struct {
	db     *gorm.DB
	casbin *casbinx.Service
}

func NewRoleBiz(db *gorm.DB, casbinService *casbinx.Service) *RoleBiz {
	return &RoleBiz{db: db, casbin: casbinService}
}

func (b *RoleBiz) List(req *api.RoleListRequest) (*api.PageResponse[*api.RoleResponse], error) {
	roles, total, err := queryPage[model.Role](b.db.Model(&model.Role{}).Order("id ASC"), req.Page, req.Limit)
	if err != nil {
		return nil, err
	}
	result := make([]*api.RoleResponse, 0, len(roles))
	for _, role := range roles {
		result = append(result, toRoleResponse(role))
	}
	return &api.PageResponse[*api.RoleResponse]{List: result, Total: total}, nil
}

func (b *RoleBiz) Create(req *api.RoleRequest) (*api.RoleResponse, error) {
	role := model.Role{Code: req.Code, Name: req.Name, Description: req.Description}
	if err := b.db.Create(&role).Error; err != nil {
		return nil, err
	}
	return toRoleResponse(role), nil
}

func (b *RoleBiz) Update(req *api.RoleRequest) (*api.RoleResponse, error) {
	if req.ID == 0 {
		return nil, ErrInvalid
	}
	var role model.Role
	if err := b.db.First(&role, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	if req.Code != role.Code {
		return nil, errors.New("role code cannot be changed")
	}
	role.Name = req.Name
	role.Description = req.Description
	if err := b.db.Save(&role).Error; err != nil {
		return nil, err
	}
	return toRoleResponse(role), nil
}

func (b *RoleBiz) Delete(req *api.RoleIDRequest) error {
	var role model.Role
	if err := b.db.First(&role, req.ID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrNotFound
		}
		return err
	}
	if role.Code == "admin" {
		return errors.New("admin role cannot be deleted")
	}
	if err := b.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role_id = ?", role.ID).Delete(&model.UserRole{}).Error; err != nil {
			return err
		}
		return tx.Delete(&role).Error
	}); err != nil {
		return err
	}
	return b.casbin.DeleteRole(role.Code)
}

func (b *RoleBiz) Catalog() []*api.Permission {
	items := casbinx.Catalog()
	result := make([]*api.Permission, 0, len(items))
	for _, item := range items {
		result = append(result, toPermission(item))
	}
	return result
}

func (b *RoleBiz) GetPermissions(req *api.PermissionRequest) ([]*api.Permission, error) {
	role, err := b.findRole(req.ID)
	if err != nil {
		return nil, err
	}
	policies := b.casbin.GetRolePolicies(role.Code)
	result := make([]*api.Permission, 0, len(policies))
	for _, policy := range policies {
		if len(policy) < 3 {
			continue
		}
		result = append(result, &api.Permission{Path: policy[1], Method: policy[2], Key: policy[2] + " " + policy[1]})
	}
	return result, nil
}

func (b *RoleBiz) SetPermissions(req *api.PermissionRequest) ([]*api.Permission, error) {
	role, err := b.findRole(req.ID)
	if err != nil {
		return nil, err
	}
	if role.Code == "admin" {
		return nil, errors.New("admin role keeps full access")
	}
	catalog := make(map[string]casbinx.Permission)
	for _, item := range casbinx.Catalog() {
		catalog[item.Key] = item
	}
	clean := make([]casbinx.Permission, 0, len(req.Permissions))
	seen := make(map[string]bool)
	for _, item := range req.Permissions {
		key := item.Method + " " + item.Path
		value, ok := catalog[key]
		if !ok {
			return nil, errors.New("unknown permission: " + key)
		}
		if !seen[key] {
			clean = append(clean, value)
			seen[key] = true
		}
	}
	if err := b.casbin.SetRolePolicies(role.Code, clean); err != nil {
		return nil, err
	}
	result := make([]*api.Permission, 0, len(clean))
	for _, item := range clean {
		result = append(result, toPermission(item))
	}
	return result, nil
}

func (b *RoleBiz) findRole(id int64) (*model.Role, error) {
	var role model.Role
	if err := b.db.First(&role, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return &role, nil
}

func toRoleResponse(role model.Role) *api.RoleResponse {
	return &api.RoleResponse{ID: role.ID, Code: role.Code, Name: role.Name, Description: role.Description, CreatedAt: role.CreatedAt, UpdatedAt: role.UpdatedAt}
}

func toPermission(permission casbinx.Permission) *api.Permission {
	return &api.Permission{Group: permission.Group, Name: permission.Name, Path: permission.Path, Method: permission.Method, Key: permission.Key}
}
