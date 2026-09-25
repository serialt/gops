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
