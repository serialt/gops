package rbac

import (
	"context"
	"gops/internal/data/define"
	"gops/internal/data/model"
	"gops/internal/data/query"

	"gorm.io/gorm"
)

type RbacRepo struct {
	db *gorm.DB
}

func NewRbacRepo(db *gorm.DB) *RbacRepo {
	return &RbacRepo{db: db}
}

func (r *RbacRepo) GetUserListByCond(ctx context.Context, cond *define.RbacCond) (items []*model.User, err error) {
	mm := query.Use(r.db).User
	role := query.Use(r.db).Role

	do := mm.WithContext(ctx)

	if len(cond.UserIds) > 0 {
		do = do.Where(mm.ID.In(cond.UserIds...))
	}
	if len(cond.RoleIds) > 0 {
		do = do.Preload(mm.Roles.On(role.ID.In(cond.RoleIds...)))
	}
	items, err = do.Order(mm.ID.Desc()).Find()
	return
}
