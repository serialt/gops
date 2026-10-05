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
	host := g.GenerateModel("host",
		gen.FieldType("port", "*int64"),
		gen.FieldType("credential_id", "*int64"),
		gen.FieldType("vnc_port", "*int64"),
		gen.FieldType("vnc_credential_id", "*int64"),
	)
	dbInstance := g.GenerateModel("db_instance")
	terminalSession := g.GenerateModel("terminal_session")
	monitorDomain := g.GenerateModel("monitor_domain")
	monitorCheckHistory := g.GenerateModel("monitor_check_history")
	monitorAlertChannel := g.GenerateModel("monitor_alert_channel")
	monitorAlertReceiver := g.GenerateModel("monitor_alert_receiver")
	monitorAlertLog := g.GenerateModel("monitor_alert_log")
	taskTemplate := g.GenerateModel("task_template")
	taskJob := g.GenerateModel("task_job")
	taskExecution := g.GenerateModel("task_execution")
	sslCertificate := g.GenerateModel("ssl_certificate")
	sslDNSProvider := g.GenerateModel("ssl_dns_provider")
	sslDeployConfig := g.GenerateModel("ssl_deploy_config")
	sslTask := g.GenerateModel("ssl_task")
	auditOperationLog := g.GenerateModel("audit_operation_log")
	auditLoginLog := g.GenerateModel("audit_login_log")
	kubernetesCluster := g.GenerateModel("kubernetes_cluster")
	kubernetesTerminalSession := g.GenerateModel("kubernetes_terminal_session")
	kubernetesClusterInspection := g.GenerateModel("kubernetes_cluster_inspection")
	kubernetesUserKubeConfig := g.GenerateModel("kubernetes_user_kube_config")
	kubernetesUserRoleBinding := g.GenerateModel("kubernetes_user_role_binding")
	iotACLRule := g.GenerateModel("iot_acl_rule")
	iotTopic := g.GenerateModel("iot_topic")
	iotReceivedMessage := g.GenerateModel("iot_received_message")
	iotClientConnection := g.GenerateModel("iot_client_connection")
	iotExternalBroker := g.GenerateModel("iot_external_broker")
	userRole := g.GenerateModel("user_role")
	userMFA := g.GenerateModel("user_mfa")
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
		userMFA,
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
		taskTemplate,
		taskJob,
		taskExecution,
		sslCertificate,
		sslDNSProvider,
		sslDeployConfig,
		sslTask,
		auditOperationLog, auditLoginLog,
		kubernetesCluster,
		kubernetesTerminalSession,
		kubernetesClusterInspection,
		kubernetesUserKubeConfig,
		kubernetesUserRoleBinding,
		iotACLRule,
		iotTopic,
		iotReceivedMessage,
		iotClientConnection,
		iotExternalBroker,
	)

}
