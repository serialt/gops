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

func (r *RbacRepo) GetUserListByCond(ctx context.Context, cond *define.UserListCond) (items []*model.User, err error) {
	mm := query.Use(r.db).User

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
