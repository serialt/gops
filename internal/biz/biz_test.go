package biz

import (
	"context"
	"gops/internal/config"
	"gops/internal/data"
	"gops/internal/data/repo/rbac"
	"gops/internal/pkg/casbinx"

	"github.com/serialt/sugar/v3"
	"gorm.io/gorm"
)

var (
	db     *gorm.DB
	ctxbiz = context.Background()

	userBiz *UserBiz
)

func init() {
	log := sugar.New(
		sugar.WithShort(true),
		sugar.WithLevel("debug"),
	)
	cfg := config.NewConfig("../../config.yaml")
	cfg.Dsn = "../../gops.db"
	db2, cleanup, err := data.NewDB(cfg, log)
	if err != nil {
		cleanup()
		panic(err)
	}
	db = db2
	service, err := casbinx.NewService(db)
	if err != nil {
		cleanup()
		panic(err)
	}
	repo := rbac.NewRbacRepo(db)
	userBiz = NewUserBiz(db, service, repo)

}
