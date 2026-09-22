package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gops/internal/biz"
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

func RegisterRoute(ctx context.Context, r *gin.Engine, db *gorm.DB, cfg *config.Config) (*gin.Engine, error) {

	casbinService, err := casbinx.New(db)
	if err != nil {
		return nil, fmt.Errorf("initialize casbin: %w", err)
	}
	if err := seed(db, casbinService); err != nil {
		return nil, fmt.Errorf("seed data: %w", err)
	}

	jwtManager := auth.NewJWTManager(cfg.Server.JwtSecret, cfg.Server.JwtTimeout)
	authBiz := biz.NewAuthBiz(db, jwtManager, casbinService)
	authHandler := handler.NewAuthHandler(authBiz)
	dashboardHandler := handler.NewDashboardHandler(biz.NewDashboardBiz(db))
	userHandler := handler.NewUserHandler(biz.NewUserBiz(db, casbinService))
	roleHandler := handler.NewRoleHandler(biz.NewRoleBiz(db, casbinService))
	organizationHandler := handler.NewOrganizationHandler(biz.NewOrganizationBiz(db))
	assetHandler := handler.NewAssetHandler(biz.NewAssetBiz(db, cfg.Server.AESKey))
	systemConfigHandler := handler.NewSystemConfigHandler(biz.NewSystemConfigBiz(db, cfg.Server.AESKey))
	monitorBiz := biz.NewMonitorBiz(db, cfg.Server.AESKey)
	monitorHandler := handler.NewMonitorHandler(monitorBiz)
	go monitorBiz.StartScheduler(ctx)
	nginxHandler := handler.NewNginxHandler(biz.NewNginxBiz(db))
	taskHandler := handler.NewTaskHandler(biz.NewTaskBiz(db))
	sslCertHandler := handler.NewSSLCertHandler(biz.NewSSLCertBiz(db))
	kubernetesBiz := biz.NewKubernetesBiz(db)
	kubernetesHandler := handler.NewKubernetesHandler(kubernetesBiz)
	go kubernetesBiz.StartStatusScheduler(ctx, 6*time.Hour)
	auditHandler := handler.NewAuditHandler(biz.NewAuditBiz(db))

	r.Use(gin.Logger(), gin.Recovery())

	r.POST("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	api.POST("/auth/login", authHandler.Login)

	protected := api.Group("")
	protected.Use(middleware.Auth(jwtManager))
	protected.POST("/auth/me", authHandler.Me)
	terminal := api.Group("")
	terminal.Use(middleware.WebSocketAuth(jwtManager), middleware.Casbin(casbinService))
	terminal.GET("/assets/terminal/connect/:id", assetHandler.ConnectTerminal)

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
	authorized.POST("/assets/groups/list", assetHandler.ListGroups)
	authorized.POST("/assets/groups/create", assetHandler.CreateGroup)
	authorized.POST("/assets/groups/update", assetHandler.UpdateGroup)
	authorized.POST("/assets/groups/delete", assetHandler.DeleteGroup)
	authorized.POST("/assets/credentials/list", assetHandler.ListCredentials)
	authorized.POST("/assets/credentials/create", assetHandler.CreateCredential)
	authorized.POST("/assets/credentials/update", assetHandler.UpdateCredential)
	authorized.POST("/assets/credentials/delete", assetHandler.DeleteCredential)
	authorized.POST("/assets/cloud-accounts/list", assetHandler.ListCloudAccounts)
	authorized.POST("/assets/cloud-accounts/options", assetHandler.CloudProviderOptions)
	authorized.POST("/assets/cloud-accounts/create", assetHandler.CreateCloudAccount)
	authorized.POST("/assets/cloud-accounts/update", assetHandler.UpdateCloudAccount)
	authorized.POST("/assets/cloud-accounts/delete", assetHandler.DeleteCloudAccount)
	authorized.POST("/assets/hosts/list", assetHandler.ListHosts)
	authorized.POST("/assets/hosts/create", assetHandler.CreateHost)
	authorized.POST("/assets/hosts/update", assetHandler.UpdateHost)
	authorized.POST("/assets/hosts/delete", assetHandler.DeleteHost)
	authorized.POST("/assets/db-instances/options", assetHandler.DatabaseTypeOptions)
	authorized.POST("/assets/db-instances/list", assetHandler.ListDBInstances)
	authorized.POST("/assets/db-instances/create", assetHandler.CreateDBInstance)
	authorized.POST("/assets/db-instances/update", assetHandler.UpdateDBInstance)
	authorized.POST("/assets/db-instances/delete", assetHandler.DeleteDBInstance)
	authorized.POST("/assets/terminal-sessions/list", assetHandler.ListTerminalSessions)
	authorized.POST("/assets/terminal-sessions/recording", assetHandler.GetTerminalSessionRecording)
	authorized.POST("/assets/terminal-sessions/delete", assetHandler.DeleteTerminalSession)
	authorized.POST("/system-config/get", systemConfigHandler.Get)
	authorized.POST("/system-config/update", systemConfigHandler.Update)
	authorized.POST("/system-config/email/test", systemConfigHandler.TestEmail)
	authorized.POST("/monitor/domains/list", monitorHandler.ListDomains)
	authorized.POST("/monitor/domains/create", monitorHandler.CreateDomain)
	authorized.POST("/monitor/domains/update", monitorHandler.UpdateDomain)
	authorized.POST("/monitor/domains/delete", monitorHandler.DeleteDomain)
	authorized.POST("/monitor/domains/check", monitorHandler.CheckDomain)
	authorized.POST("/monitor/domains/history", monitorHandler.ListCheckHistory)
	authorized.POST("/monitor/domains/stats", monitorHandler.DomainStats)
	authorized.POST("/monitor/channels/list", monitorHandler.ListChannels)
	authorized.POST("/monitor/channels/save", monitorHandler.SaveChannel)
	authorized.POST("/monitor/channels/delete", monitorHandler.DeleteChannel)
	authorized.POST("/monitor/channels/test", monitorHandler.TestChannel)
	authorized.POST("/monitor/receivers/list", monitorHandler.ListReceivers)
	authorized.POST("/monitor/receivers/save", monitorHandler.SaveReceiver)
	authorized.POST("/monitor/receivers/delete", monitorHandler.DeleteReceiver)
	authorized.POST("/monitor/logs/list", monitorHandler.ListLogs)
	authorized.POST("/nginx/sources/list", nginxHandler.ListSources)
	authorized.POST("/nginx/sources/save", nginxHandler.SaveSource)
	authorized.POST("/nginx/sources/delete", nginxHandler.DeleteSource)
	authorized.POST("/nginx/overview", nginxHandler.Overview)
	authorized.POST("/nginx/logs/list", nginxHandler.ListLogs)
	authorized.POST("/tasks/jobs/list", taskHandler.ListJobs)
	authorized.POST("/tasks/jobs/save", taskHandler.SaveJob)
	authorized.POST("/ssl-cert/certificates/list", sslCertHandler.ListCertificates)
	authorized.POST("/ssl-cert/certificates/save", sslCertHandler.SaveCertificate)
	authorized.POST("/kubernetes/clusters/list", kubernetesHandler.ListClusters)
	authorized.POST("/kubernetes/clusters/save", kubernetesHandler.SaveCluster)
	authorized.POST("/kubernetes/clusters/delete", kubernetesHandler.DeleteCluster)
	authorized.POST("/kubernetes/clusters/test", kubernetesHandler.TestCluster)
	authorized.POST("/kubernetes/clusters/sync", kubernetesHandler.SyncClusters)
	authorized.POST("/kubernetes/clusters/config", kubernetesHandler.GetClusterConfig)
	authorized.POST("/kubernetes/nodes/list", kubernetesHandler.ListNodes)
	authorized.POST("/kubernetes/nodes/update", kubernetesHandler.UpdateNode)
	authorized.POST("/kubernetes/nodes/delete", kubernetesHandler.DeleteNode)
	authorized.POST("/kubernetes/nodes/drain", kubernetesHandler.DrainNode)
	authorized.POST("/kubernetes/namespaces/list", kubernetesHandler.ListNamespaces)
	authorized.POST("/kubernetes/namespaces/update", kubernetesHandler.UpdateNamespace)
	authorized.POST("/kubernetes/workloads/list", kubernetesHandler.ListWorkloads)
	authorized.POST("/kubernetes/workloads/stats", kubernetesHandler.WorkloadStats)
	authorized.POST("/kubernetes/pods/list", kubernetesHandler.ListPods)
	authorized.POST("/kubernetes/stats/get", kubernetesHandler.GetClusterStats)
	authorized.POST("/kubernetes/events/list", kubernetesHandler.ListEvents)
	authorized.POST("/kubernetes/resources/list", kubernetesHandler.ListKubernetesObjects)
	authorized.POST("/kubernetes/resources/get", kubernetesHandler.GetKubernetesObject)
	authorized.POST("/kubernetes/resources/template", kubernetesHandler.GetKubernetesResourceTemplate)
	authorized.POST("/kubernetes/resources/apply", kubernetesHandler.ApplyKubernetesObject)
	authorized.POST("/kubernetes/resources/delete", kubernetesHandler.DeleteKubernetesObject)
	authorized.POST("/kubernetes/resources/labels/update", kubernetesHandler.UpdateKubernetesObjectLabels)
	authorized.POST("/kubernetes/inspections/start", kubernetesHandler.StartInspection)
	authorized.POST("/kubernetes/inspections/progress", kubernetesHandler.GetInspectionProgress)
	authorized.POST("/kubernetes/inspections/result", kubernetesHandler.GetInspectionResult)
	authorized.POST("/kubernetes/inspections/history", kubernetesHandler.ListInspectionHistory)
	authorized.POST("/kubernetes/inspections/delete", kubernetesHandler.DeleteInspection)
	authorized.POST("/audit/operation-logs/list", auditHandler.ListOperationLogs)
	authorized.POST("/audit/login-logs/list", auditHandler.ListLoginLogs)
	authorized.POST("/audit/data-logs/list", auditHandler.ListDataLogs)

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
	schedulerCtx, cancelSchedulers := context.WithCancel(context.Background())
	whitelistMap := lo.KeyBy(whitelist, func(item string) string { return item })
	middleware.RegisterMiddleware(r, cfg, whitelistMap)
	if _, err := RegisterRoute(schedulerCtx, r, db, cfg); err != nil {
		cancelSchedulers()
		panic(err)
	}
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
