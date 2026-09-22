package casbinx

import (
	"strconv"

	"github.com/casbin/casbin/v3"
	casbinmodel "github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"gorm.io/gorm"
)

type Service struct {
	Enforcer *casbin.Enforcer
}

func NewService(db *gorm.DB) (*Service, error) {
	adapter, err := gormadapter.NewAdapterByDB(db)
	if err != nil {
		return nil, err
	}

	m, err := casbinmodel.NewModelFromString(modelText)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer(m, adapter)
	if err != nil {
		return nil, err
	}
	if err := enforcer.LoadPolicy(); err != nil {
		return nil, err
	}

	return &Service{Enforcer: enforcer}, nil
}

func (s *Service) SetUserRoles(userID uint, roleCodes []string) error {
	subject := strconv.FormatUint(uint64(userID), 10)
	if _, err := s.Enforcer.DeleteRolesForUser(subject); err != nil {
		return err
	}
	for _, code := range roleCodes {
		if _, err := s.Enforcer.AddGroupingPolicy(subject, code); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) SetRolePolicies(roleCode string, permissions []Permission) error {
	if _, err := s.Enforcer.RemoveFilteredPolicy(0, roleCode); err != nil {
		return err
	}
	for _, item := range permissions {
		if _, err := s.Enforcer.AddPolicy(roleCode, item.Path, item.Method); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) GetRolePolicies(roleCode string) [][]string {
	resp, _ := s.Enforcer.GetFilteredPolicy(0, roleCode)
	return resp

}

func (s *Service) DeleteRole(roleCode string) error {
	if _, err := s.Enforcer.RemoveFilteredPolicy(0, roleCode); err != nil {
		return err
	}
	_, err := s.Enforcer.RemoveFilteredGroupingPolicy(1, roleCode)
	return err
}

func (s *Service) AllowedPermissionKeys(userID uint) []string {
	subject := strconv.FormatUint(uint64(userID), 10)
	result := make([]string, 0)
	for _, item := range Catalog() {
		allowed, err := s.Enforcer.Enforce(subject, item.Path, item.Method)
		if err == nil && allowed {
			result = append(result, item.Key)
		}
	}
	return result
}
