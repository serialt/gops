package server

import (
	"fmt"
	"net/http"

	"gops/internal/config"
	"gops/internal/handler"
	"gops/internal/middleware"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	NewHTTPServer,
	NewGinEngine,
)

var whitelist = []string{
	"/token",
	"/api/v1/auth/login",
	"/api/v1/user/create",
	"/api/v1/public/login",
	"/api/v1/public/register",
	"/health",
	"/api/v1/public/config",
	"/api/v1/captcha",
	"/api/v1/captcha/verify",
	"/api/v1/auth/mfa",
	"/api/v1/auth/mfa/login",
}

func RegisterRoute(r *gin.Engine, db *gorm.DB, cfg *config.Config) (*gin.Engine, error) {

	casbinService, err := casbinx.New(db)
	if err != nil {
		return nil, fmt.Errorf("initialize casbin: %w", err)
	}
	if err := seed(db, casbinService, cfg.Server.AdminPass); err != nil {
		return nil, fmt.Errorf("seed data: %w", err)
	}

	jwtManager := auth.NewJWTManager(cfg.Server.JwtSecret, cfg.Server.JwtTimeout)
	authHandler := handler.NewAuthHandler(db, jwtManager, casbinService)
	dashboardHandler := handler.NewDashboardHandler(db)
	userHandler := handler.NewUserHandler(db, casbinService)
	roleHandler := handler.NewRoleHandler(db, casbinService)
	organizationHandler := handler.NewOrganizationHandler(db)

	r.Use(gin.Logger(), gin.Recovery())

	r.POST("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	protected.POST("/auth/me", authHandler.Me)

	authorized := protected.Group("")
	authorized.Use(middleware.Casbin(casbinService))
	authorized.POST("/dashboard/stats", dashboardHandler.Stats)
	authorized.POST("/users/list", userHandler.List)
	authorized.POST("/users/create", userHandler.Create)
	authorized.POST("/users/update", userHandler.Update)
	authorized.POST("/users/delete", userHandler.Delete)
	authorized.POST("/roles/list", roleHandler.List)
	authorized.POST("/roles/create", roleHandler.Create)
	authorized.POST("/roles/update", roleHandler.Update)
	authorized.POST("/roles/delete", roleHandler.Delete)
	authorized.POST("/permissions/catalog", roleHandler.Catalog)
	authorized.POST("/roles/permissions/get", roleHandler.GetPermissions)
	authorized.POST("/roles/permissions/set", roleHandler.SetPermissions)
	authorized.POST("/organizations/list", organizationHandler.List)
	authorized.POST("/organizations/create", organizationHandler.Create)
	authorized.POST("/organizations/update", organizationHandler.Update)
	authorized.POST("/organizations/delete", organizationHandler.Delete)
	authorized.POST("/organizations/move", organizationHandler.Move)

	return r, nil
}

func NewGinEngine(
	cfg *config.Config,
	db *gorm.DB,
) (r *gin.Engine, cleanup func()) {

	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r = gin.New()
	whitelistMap := lo.KeyBy(whitelist, func(item string) string { return item })
	middleware.RegisterMiddleware(r, cfg, whitelistMap)
	RegisterRoute(r, db, cfg)
	return
}

func NewHTTPServer(engine *gin.Engine, cfg *config.Config) (*http.Server, func(), error) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HttpPort),
		Handler: engine,
	}
	cleanup := func() {}
	return srv, cleanup, nil
}
