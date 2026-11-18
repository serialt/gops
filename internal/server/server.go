package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"gops/internal/biz"
	"gops/internal/config"
	"gops/internal/handler"
	iotbroker "gops/internal/iot"
	"gops/internal/middleware"
	"gops/internal/pkg/auth"
	"gops/internal/pkg/casbinx"

	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/samber/lo"
	"gorm.io/gorm"
)

var ProviderSet = wire.NewSet(
	casbinx.NewService,
	auth.NewJWTManager,
	handler.NewAuthHandler,
	handler.NewDashboardHandler,
	handler.NewUserHandler,
	handler.NewRoleHandler,
	handler.NewOrganizationHandler,
	handler.NewAssetHandler,
	handler.NewSystemConfigHandler,
	handler.NewMonitorHandler,
	handler.NewTaskHandler,
	handler.NewSSLCertHandler,
	handler.NewKubernetesHandler,
	handler.NewAuditHandler,
	handler.NewIOTHandler,
	iotbroker.NewBroker,
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

func NewGinEngine(
	cfg *config.Config,
	db *gorm.DB,
	casbinService *casbinx.Service,
	jwtManager *auth.JWTManager,
	authHandler *handler.AuthHandler,
	dashboardHandler *handler.DashboardHandler,
	userHandler *handler.UserHandler,
	roleHandler *handler.RoleHandler,
	organizationHandler *handler.OrganizationHandler,
	assetHandler *handler.AssetHandler,
	systemConfigHandler *handler.SystemConfigHandler,
	monitorHandler *handler.MonitorHandler,
	taskHandler *handler.TaskHandler,
	sslCertHandler *handler.SSLCertHandler,
	kubernetesHandler *handler.KubernetesHandler,
	auditHandler *handler.AuditHandler,
	iotHandler *handler.IOTHandler,
	metricsBiz *biz.MetricsBiz,
	iotBroker *iotbroker.Broker,
) (r *gin.Engine, cleanup func()) {

	if cfg.Log.Level != "debug" && cfg.Log.Level != "dev" {
		gin.SetMode(gin.ReleaseMode)
	}

	r = gin.New()
	r.Use(gzipMiddleware())
	schedulerCtx, cancelSchedulers := context.WithCancel(context.Background())
	whitelistMap := lo.KeyBy(whitelist, func(item string) string { return item })
	middleware.RegisterMiddleware(r, cfg, whitelistMap)

	// if cfg.Env == define.EnvDev {
	// 	if err := Seed(db, casbinService); err != nil {
	// 		cancelSchedulers()
	// 		panic(err)
	// 	}
	// }

	go monitorHandler.StartScheduler(schedulerCtx)
	go assetHandler.StartAssetCollectionScheduler(schedulerCtx)
	go kubernetesHandler.StartStatusScheduler(schedulerCtx, 6*time.Hour)
	go func() {
		if err := iotBroker.Run(); err != nil {
			fmt.Printf("IOT MQTT broker stopped: %v\n", err)
		}
	}()

	r.Use(gin.Logger(), gin.Recovery())

	var metricsServer *http.Server
	if cfg.Server.Metrics.Enabled {
		metricsRegistry := metricsBiz.NewPrometheusRegistry(biz.NewMetricsCollector(metricsBiz))

		metricsEngine := gin.New()
		metricsEngine.Use(gin.Recovery())
		metricsEngine.GET("/metrics", gin.WrapH(promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{})))
		metricsEngine.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		metricsServer = &http.Server{
			Addr:    fmt.Sprintf(":%d", cfg.Server.Metrics.HTTPPort),
			Handler: metricsEngine,
		}
		go func() {
			log.Printf("Running metrics server at :%d", cfg.Server.Metrics.HTTPPort)
			if err := metricsServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Printf("metrics server failed: %v", err)
			}
		}()
	}

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
	terminal.GET("/assets/terminal/vnc/connect/:id", assetHandler.ConnectVNCTerminal)
	terminal.GET("/assets/terminal/database/connect/:id", assetHandler.ConnectDatabaseTerminal)

	authorized := protected.Group("")
	authorized.Use(middleware.Casbin(casbinService), middleware.OperationAudit(db))
	dashboardHandler.RegisterRoutes(authorized)
	userHandler.RegisterRoutes(authorized)
	roleHandler.RegisterRoutes(authorized)
	organizationHandler.RegisterRoutes(authorized)
	assetHandler.RegisterRoutes(authorized)
	systemConfigHandler.RegisterRoutes(authorized)
	monitorHandler.RegisterRoutes(authorized)
	taskHandler.RegisterRoutes(authorized)
	sslCertHandler.RegisterRoutes(authorized)
	kubernetesHandler.RegisterRoutes(authorized)
	auditHandler.RegisterRoutes(authorized)
	iotHandler.RegisterRoutes(authorized)

	registerFrontend(r, db)
	cleanup = func() {
		cancelSchedulers()
		_ = iotBroker.Close()
		if metricsServer != nil {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_ = metricsServer.Shutdown(ctx)
		}
	}
	return
}

func NewHTTPServer(engine *gin.Engine, cfg *config.Config) (*http.Server, func(), error) {
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.HttpPort),
		Handler: engine,
	}
	cleanup := func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(ctx)
	}
	return srv, cleanup, nil
}
