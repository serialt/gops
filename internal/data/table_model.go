package data

import (
	"gorm.io/gen"
	"gorm.io/gen/field"
)

func gopsTable(g *gen.Generator) {
	role := g.GenerateModel("role")
	organization := g.GenerateModel("organization")
	organizationUser := g.GenerateModel("organization_user")
	systemConfig := g.GenerateModel("system_config")
	assetGroup := g.GenerateModel("asset_group")
	credential := g.GenerateModel("credential")
	cloudAccount := g.GenerateModel("cloud_account")
	host := g.GenerateModel("host")
	dbInstance := g.GenerateModel("db_instance")
	terminalSession := g.GenerateModel("terminal_session")
	monitorDomain := g.GenerateModel("monitor_domain")
	monitorCheckHistory := g.GenerateModel("monitor_check_history")
	monitorAlertChannel := g.GenerateModel("monitor_alert_channel")
	monitorAlertReceiver := g.GenerateModel("monitor_alert_receiver")
	monitorAlertLog := g.GenerateModel("monitor_alert_log")
	nginxSource := g.GenerateModel("nginx_source")
	nginxAccessLog := g.GenerateModel("nginx_access_log")
	taskTemplate := g.GenerateModel("task_template")
	taskJob := g.GenerateModel("task_job")
	taskExecution := g.GenerateModel("task_execution")
	sslCertificate := g.GenerateModel("ssl_certificate")
	sslDNSProvider := g.GenerateModel("ssl_dns_provider")
	auditOperationLog := g.GenerateModel("audit_operation_log")
	auditLoginLog := g.GenerateModel("audit_login_log")
	auditDataLog := g.GenerateModel("audit_data_log")
	kubernetesCluster := g.GenerateModel("kubernetes_cluster")
	userRole := g.GenerateModel("user_role")
	user := g.GenerateModel("user", gen.FieldRelate(field.HasMany, "Roles", role,
		&field.RelateConfig{
			RelateSlicePointer: true,
			GORMTag: field.GormTag{
				"many2many": []string{
					"user_role",
				},
			},
		},
	), gen.FieldRelate(field.HasMany, "Organizations", organization,
		&field.RelateConfig{RelateSlicePointer: true, GORMTag: field.GormTag{"many2many": []string{"organization_user"}}},
	))
	g.ApplyBasic(
		role,
		userRole,
		user,
		organization,
		organizationUser,
		systemConfig,
		assetGroup,
		credential,
		cloudAccount,
		host,
		dbInstance,
		terminalSession,
		monitorDomain,
		monitorCheckHistory,
		monitorAlertChannel,
		monitorAlertReceiver,
		monitorAlertLog,
		nginxSource,
		nginxAccessLog,
		taskTemplate,
		taskJob,
		taskExecution,
		sslCertificate,
		sslDNSProvider,
		auditOperationLog, auditLoginLog, auditDataLog,
		kubernetesCluster,
	)

}
