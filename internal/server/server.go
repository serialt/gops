package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
	auth.NewJWTManager,
	handler.NewAuthHandler,
	handler.NewDashboardHandler,
	handler.NewUserHandler,
	handler.NewRoleHandler,
	handler.NewOrganizationHandler,
	handler.NewAssetHandler,
	handler.NewSystemConfigHandler,
	handler.NewMonitorHandler,
	handler.NewNginxHandler,
	handler.NewTaskHandler,
	handler.NewSSLCertHandler,
	handler.NewKubernetesHandler,
	handler.NewAuditHandler,
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

func RegisterRoute(ctx context.Context, r *gin.Engine, db *gorm.DB, cfg *config.Config,
	jwtManager *auth.JWTManager,
	authHandler *handler.AuthHandler,
	dashboardHandler *handler.DashboardHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	organizationHandler *handler.OrganizationHandler,
	assetHandler *handler.AssetHandler,
	systemConfigHandler *handler.SystemConfigHandler,
	monitorHandler *handler.MonitorHandler,
	nginxHandler *handler.NginxHandler,
	taskHandler *handler.TaskHandler,
	sslCertHandler *handler.SSLCertHandler,
	kubernetesHandler *handler.KubernetesHandler,
	auditHandler *handler.AuditHandler,
) (*gin.Engine, error) {

	casbinService, err := casbinx.New(db)
	if err != nil {
		return nil, fmt.Errorf("initialize casbin: %w", err)
	}
	if err := seed(db, casbinService); err != nil {
		return nil, fmt.Errorf("seed data: %w", err)
	}

	// go monitorBiz.StartScheduler(ctx)

	// go kubernetesBiz.StartStatusScheduler(ctx, 6*time.Hour)

	r.Use(gin.Logger(), gin.Recovery())

	r.POST("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	authHandler.RegisterRoutesNologin(api)
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	authHandler.RegisterRoutes(protected)

	terminal := api.Group("")
	terminal.Use(middleware.WebSocketAuth(jwtManager), middleware.Casbin(casbinService))
	terminal.GET("/assets/terminal/connect/:id", assetHandler.ConnectTerminal)

	authorized := protected.Group("")
	authorized.Use(middleware.Casbin(casbinService))
	dashboardHandler.RegisterRoutes(authorized)
	userHandler.RegisterRoutes(authorized)
	roleHandler.RegisterRoutes(authorized)
	organizationHandler.RegisterRoutes(authorized)
	assetHandler.RegisterRoutes(authorized)
	systemConfigHandler.RegisterRoutes(authorized)
	monitorHandler.RegisterRoutes(authorized)
	nginxHandler.RegisterRoutes(authorized)
	taskHandler.RegisterRoutes(authorized)
	sslCertHandler.RegisterRoutes(authorized)
	kubernetesHandler.RegisterRoutes(authorized)
	auditHandler.RegisterRoutes(authorized)

	return r, nil
}

func NewGinEngine(
	cfg *config.Config,
	db *gorm.DB,
	jwtManager *auth.JWTManager,
	authHandler *handler.AuthHandler,
	dashboardHandler *handler.DashboardHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	organizationHandler *handler.OrganizationHandler,
	assetHandler *handler.AssetHandler,
	systemConfigHandler *handler.SystemConfigHandler,
	monitorHandler *handler.MonitorHandler,
	nginxHandler *handler.NginxHandler,
	taskHandler *handler.TaskHandler,
	sslCertHandler *handler.SSLCertHandler,
	kubernetesHandler *handler.KubernetesHandler,
	auditHandler *handler.AuditHandler,
) (r *gin.Engine, cleanup func()) {

	if cfg.Log.Level != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	r = gin.New()
	schedulerCtx, cancelSchedulers := context.WithCancel(context.Background())
	whitelistMap := lo.KeyBy(whitelist, func(item string) string { return item })
	middleware.RegisterMiddleware(r, cfg, whitelistMap)

	casbinService, err := casbinx.New(db)
	if err != nil {
		cancelSchedulers()
		panic(err)
	}
	if err := seed(db, casbinService); err != nil {
		cancelSchedulers()
		panic(err)
	}

	go monitorHandler.StartScheduler(schedulerCtx)
	go kubernetesHandler.StartStatusScheduler(schedulerCtx, 6*time.Hour)

	r.Use(gin.Logger(), gin.Recovery())

	r.POST("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")

	authHandler.RegisterRoutesNologin(api)
	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	authHandler.RegisterRoutes(protected)

	terminal := api.Group("")
	terminal.Use(middleware.WebSocketAuth(jwtManager), middleware.Casbin(casbinService))
	terminal.GET("/assets/terminal/connect/:id", assetHandler.ConnectTerminal)

	authorized := protected.Group("")
	authorized.Use(middleware.Casbin(casbinService))
	dashboardHandler.RegisterRoutes(authorized)
	userHandler.RegisterRoutes(authorized)
	roleHandler.RegisterRoutes(authorized)
	organizationHandler.RegisterRoutes(authorized)
	assetHandler.RegisterRoutes(authorized)
	systemConfigHandler.RegisterRoutes(authorized)
	monitorHandler.RegisterRoutes(authorized)
	nginxHandler.RegisterRoutes(authorized)
	taskHandler.RegisterRoutes(authorized)
	sslCertHandler.RegisterRoutes(authorized)
	kubernetesHandler.RegisterRoutes(authorized)
	auditHandler.RegisterRoutes(authorized)

	registerFrontend(r)
	cleanup = cancelSchedulers
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
