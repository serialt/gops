package rbac

import (
	"context"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gorm"
)

type RbacRepo struct{}

func NewRbacRepo() *RbacRepo { return &RbacRepo{} }

func (r *RbacRepo) GetUserListByCond(ctx context.Context, db *gorm.DB, cond *define.UserListCond) (items []*model.User, err error) {
	mm := query.Use(db).User

	do := mm.WithContext(ctx)

	if len(cond.UserIds) > 0 {
		do = do.Where(mm.ID.In(cond.UserIds...))
	}

	do.Preload(mm.Roles, mm.Organizations)

	if cond.Keyword != "" {
		do = do.Where(mm.Username.Like("%" + cond.Keyword + "%")).Or(mm.Nickname.Like("%" + cond.Keyword + "%"))
	}
	items, err = do.Order(mm.ID.Desc()).Find()
	return
}

func (r *RbacRepo) ListUsers(ctx context.Context, db *gorm.DB) ([]*model.User, error) {
	users := query.Use(db).User
	return users.WithContext(ctx).Order(users.ID.Asc()).Find()
}

func (r *RbacRepo) GetUserByUsername(ctx context.Context, db *gorm.DB, username string) (*model.User, error) {
	users := query.Use(db).User
	return users.WithContext(ctx).Where(users.Username.Eq(username)).First()
}

func (r *RbacRepo) CreateUser(ctx context.Context, db *gorm.DB, user *model.User) error {
	return query.Use(db).User.WithContext(ctx).Create(user)
}

func (r *RbacRepo) UpdateUserPassword(ctx context.Context, db *gorm.DB, userID int64, passwordHash string) error {
	users := query.Use(db).User
	result, err := users.WithContext(ctx).Where(users.ID.Eq(userID)).UpdateSimple(users.PasswordHash.Value(passwordHash))
	if err != nil {
		return err
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *RbacRepo) GetOrganizationByCode(ctx context.Context, db *gorm.DB, code string) (*model.Organization, error) {
	organizations := query.Use(db).Organization
	return organizations.WithContext(ctx).Where(organizations.Code.Eq(code)).First()
}

func (r *RbacRepo) GetRoleByCode(ctx context.Context, db *gorm.DB, code string) (*model.Role, error) {
	roles := query.Use(db).Role
	return roles.WithContext(ctx).Where(roles.Code.Eq(code)).First()
}

func (r *RbacRepo) AppendUserOrganization(ctx context.Context, db *gorm.DB, user *model.User, organization *model.Organization) error {
	return query.Use(db).User.Organizations.WithContext(ctx).Model(user).Append(organization)
}

func (r *RbacRepo) AppendUserRole(ctx context.Context, db *gorm.DB, user *model.User, role *model.Role) error {
	return query.Use(db).User.Roles.WithContext(ctx).Model(user).Append(role)
}
