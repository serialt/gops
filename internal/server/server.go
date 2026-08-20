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

	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	protected.GET("/auth/me", authHandler.Me)

	authorized := protected.Group("")
	authorized.Use(middleware.Casbin(casbinService))
	authorized.GET("/dashboard/stats", dashboardHandler.Stats)
	authorized.GET("/users", userHandler.List)
	authorized.POST("/users", userHandler.Create)
	authorized.PUT("/users/:id", userHandler.Update)
	authorized.DELETE("/users/:id", userHandler.Delete)
	authorized.GET("/roles", roleHandler.List)
	authorized.POST("/roles", roleHandler.Create)
	authorized.PUT("/roles/:id", roleHandler.Update)
	authorized.DELETE("/roles/:id", roleHandler.Delete)
	authorized.GET("/permissions/catalog", roleHandler.Catalog)
	authorized.GET("/roles/:id/permissions", roleHandler.GetPermissions)
	authorized.PUT("/roles/:id/permissions", roleHandler.SetPermissions)

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
